package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	buildkiteAgent "github.com/buildkite/agent/agent"
	buildkite "github.com/buildkite/agent/api"
	"github.com/buildkite/agent/retry"

	"github.com/mitchellh/packer/packer"
)

type PostProcessor struct {
	client buildkiteAgent.APIClient
	jobId  string
}

func (p *PostProcessor) Configure(...interface{}) error {

	agentEndpoint := os.Getenv("BUILDKITE_AGENT_ENDPOINT")
	agentAccessToken := os.Getenv("BUILDKITE_AGENT_ACCESS_TOKEN")

	if agentEndpoint == "" || agentAccessToken == "" {
		return fmt.Errorf("must set BUILDKITE_AGENT_ENDPOINT and BUILDKITE_AGENT_ACCESS_TOKEN environment variables")
	}

	jobId := os.Getenv("BUILDKITE_JOB_ID")

	if agentEndpoint == "" || agentAccessToken == "" {
		return fmt.Errorf("must set BUILDKITE_JOB_ID environment variable")
	}

	p.client = buildkiteAgent.APIClient{
		Endpoint: agentEndpoint,
		Token:    agentAccessToken,
	}
	p.jobId = jobId

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	client := p.client.Create()

	metadata := make([]buildkite.MetaData, 0, 2)

	builderId := artifact.BuilderId()

	// packer's artifact model doesn't support AMIs very well since
	// both a region and an id need to be encoded into a single field.
	// We handle it as a special case so we can make more useful metadata
	// in the buildkite job.
	// N.B. this doesn't support AMIs being created across multiple regions.

	if builderId == "mitchellh.amazonebs" || builderId == "mitchellh.amazoninstance" || builderId == "mitchellh.amazonchroot" {
		packedId := artifact.Id()
		parts := strings.SplitN(packedId, ":", 2)
		metadata = append(metadata, buildkite.MetaData{
			Key:   "artifact_ami_id",
			Value: parts[1],
		})
		metadata = append(metadata, buildkite.MetaData{
			Key:   "artifact_ami_region",
			Value: parts[0],
		})
	} else {
		id := artifact.Id()
		if id != "" {
			metadata = append(metadata, buildkite.MetaData{
				Key:   "artifact_id",
				Value: id,
			})
		}

		// TODO: Support uploading files as BuildKite artifacts
	}

	if len(metadata) > 0 {
		ui.Message("Setting metadata in BuildKite:")
		for _, item := range metadata {
			ui.Message(fmt.Sprintf("- %s: %s", item.Key, item.Value))
			err := retry.Do(func(s *retry.Stats) error {
				resp, err := client.MetaData.Set(p.jobId, &item)
				if resp != nil && (resp.StatusCode == 401 || resp.StatusCode == 404) {
					s.Break()
				}
				if err != nil {
					log.Println(err)
				}

				return err
			}, &retry.Config{Maximum: 10, Interval: 1 * time.Second})

			if err != nil {
				return nil, false, err
			}
		}
	}

	return artifact, true, nil
}

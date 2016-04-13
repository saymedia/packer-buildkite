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

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config options
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Prefix              string `mapstructure:"prefix"`
	ctx                 interpolate.Context
}

// PostProcessor master
type PostProcessor struct {
	config Config
	client buildkiteAgent.APIClient
	jobID  string
}

// Configure sets up the config options to be used later
func (p *PostProcessor) Configure(raws ...interface{}) error {

	// Configure
	err := config.Decode(&p.config,
		&config.DecodeOpts{
			Interpolate: false,
		},
		raws...)
	if err != nil {
		return err
	}

	// apply default, var not required
	if p.config.Prefix == "" {
		p.config.Prefix = ""
	}

	agentEndpoint := os.Getenv("BUILDKITE_AGENT_ENDPOINT")
	agentAccessToken := os.Getenv("BUILDKITE_AGENT_ACCESS_TOKEN")

	if agentEndpoint == "" || agentAccessToken == "" {
		return fmt.Errorf("must set BUILDKITE_AGENT_ENDPOINT and BUILDKITE_AGENT_ACCESS_TOKEN environment variables")
	}

	jobID := os.Getenv("BUILDKITE_JOB_ID")

	if agentEndpoint == "" || agentAccessToken == "" {
		return fmt.Errorf("must set BUILDKITE_JOB_ID environment variable")
	}

	p.client = buildkiteAgent.APIClient{
		Endpoint: agentEndpoint,
		Token:    agentAccessToken,
	}
	p.jobID = jobID

	return nil
}

// PostProcess sends the Artifact to Buildkite
func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	client := p.client.Create()

	metadata := make([]buildkite.MetaData, 0, 2)

	builderID := artifact.BuilderId()

	// packer's artifact model doesn't support AMIs very well since
	// both a region and an id need to be encoded into a single field.
	// We handle it as a special case so we can make more useful metadata
	// in the buildkite job.
	// N.B. this doesn't support AMIs being created across multiple regions.

	if builderID == "mitchellh.amazonebs" || builderID == "mitchellh.amazoninstance" || builderID == "mitchellh.amazonchroot" {
		packedID := artifact.Id()
		parts := strings.SplitN(packedID, ":", 2)

		idKeyName := "artifact_ami_id"
		regionKeyName := "artifact_ami_id"

		// override key names with an optional prefixed version
		if p.config.Prefix != "" {
			idKeyName = fmt.Sprintf("%s_%s", p.config.Prefix, idKeyName)
			regionKeyName = fmt.Sprintf("%s_%s", p.config.Prefix, regionKeyName)
		}

		metadata = append(metadata, buildkite.MetaData{
			Key:   idKeyName,
			Value: parts[1],
		})
		metadata = append(metadata, buildkite.MetaData{
			Key:   regionKeyName,
			Value: parts[0],
		})
	} else {
		id := artifact.Id()

		idKeyName := "artifact_id"

		// override key name with an optional prefixed version
		if p.config.Prefix != "" {
			idKeyName = fmt.Sprintf("%s_%s", p.config.Prefix, idKeyName)
		}

		if id != "" {
			metadata = append(metadata, buildkite.MetaData{
				Key:   idKeyName,
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
				resp, err := client.MetaData.Set(p.jobID, &item)
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

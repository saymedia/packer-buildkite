packer-buildkite
================

This is an [Packer](https://packer.io/) post-processor plugin
which will write id of an artifact into the [Buildkite](https://buildkite.com/)
job metadata store.

This has been used as part of the application build pipeline at [Say Media](https://www.saymedia.com/).

Usage
-----
Add the post-processor to your packer template:

    {
        "post-processors": [
          {
            "type": "buildkite",
            "prefix": "example_a"
          }
        ]
    }

Check out [`multi-test.json`](./multi-test.json) for an example.

Installation
------------

Run:

    $ go get github.com/saymedia/packer-post-processor-buildkite
    $ go install github.com/saymedia/packer-post-processor-buildkite

Add the post-processor to ~/.packerconfig:

    {
      "post-processors": {
        "buildkite": "packer-post-processor-buildkite"
      }
    }

Tests
-----

### STEP 0:

Install packer-post-processor-buildkite as detailed above.

### STEP 1:

Create a file called `multi-test-variables.json` with the contents similar to (use your own AWS values of course):

    {
        "AWS_VPC_ID": "vpc-2421cc41",
        "AWS_SUBNET_ID": "subnet-49f75e2c",
        "AWS_SG_ID": "sg-9a57f80a",
        "AWS_AMI_ID": "ami-cb9d728f",
        "AWS_REGION": "us-west-1"
    }

### STEP 2:

Run:
  
    $ export AWS_ACCESS_KEY='<YOUR_AWS_ACCESS_KEY>'
    $ export AWS_SECRET_KEY='<YOUR_AWS_SECRET_KEY>'
    $ BUILDKITE_AGENT_ACCESS_TOKEN='<YOUR_BUILDKITE_AGENT_ACCESS_TOKEN>' BUILDKITE_JOB_ID='<YOUR_BUILDKITE_JOB_ID>' BUILDKITE_AGENT_ENDPOINT=https://agent.buildkite.com/v3 ./test.sh

License
-------

The MIT License (MIT)

Copyright (c) 2016 Say Media, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

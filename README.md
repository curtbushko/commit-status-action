### commit-status-action

A Github Action to update a commit status.

### Inputs

| Input              | Description                                               | Required             | Default |
| ------------------ | --------------------------------------------------------- | -------------------- | ------- |
| `token`       | GITHUB_TOKEN or your own token if you need to update status checks to another repo | true  | |
| `state`       | The status of the check: success, error, failure, pending or cancelled (sets status as error) | true | |
| `context`    | The context, this is displayed as the name of the check | false | default |
| `description` | Short text explaining the status of the check | false | |
| `owner`     | Repository owner | false | github.repository_owner |
| `repository` | Repository | false | github.repository |
| `sha` | SHA of the commit to update status on | false | github.sha |
| `details_url` | URL/URI to use for further details | false | |

### Running in workflows

The best way to run this action is by running the docker image directly.

```
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - name: Test commit-status-action
      uses: docker://ghcr.io/curtbushko/commit-status-action:142b02ef5528929afe4be79ec62fe9f7ad7c7ea9
      env:
        INPUT_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        INPUT_STATE: ${{ job.status }}
        INPUT_CONTEXT: "status test" 
        INPUT_DESCRIPTION: "status test"
        INPUT_OWNER: ${{ github.repository_owner }}
        INPUT_REPOSITORY: ${{ github.repository }}
        INPUT_SHA: ${{ github.event.pull_request.head.sha || github.sha }}
        INPUT_DETAILS_URL: "http://foo"
```

Where the tag for the commit-status-action image is listed [as a package in ghcr.io](https://github.com/curtbushko/commit-status-action/pkgs/container/commit-status-action)


### Running locally

1) Build the binary by running `make build`
2) Create a PAT token under [your github settings](https://github.com/settings/tokens). Export it as GITHUB_TOKEN.
3) Run the command as:

```
INPUT_TOKEN=$GITHUB_TOKEN INPUT_STATE=success INPUT_CONTEXT="status check test" INPUT_DESCRIPTION="testing.."
INPUT_OWNER="<your github account or org>" INPUT_REPOSITORY="<your github repository>" INPUT_SHA="<the SHA of a commit
on github>" INPUT_DETAILS_URL="https://foo"./bin
```

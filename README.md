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
  

### Running locally

1) Build the binary by running `make build`
2) Create a PAT token under [your github settings](https://github.com/settings/tokens). Export it as GITHUB_TOKEN.
3) Run the command as:

```
INPUT_TOKEN=$GITHUB_TOKEN INPUT_STATE=success INPUT_CONTEXT="status check test" INPUT_DESCRIPTION="testing.."
INPUT_OWNER="<your github account or org>" INPUT_REPOSITORY="<your github repository>" INPUT_SHA="<the SHA of a commit
on github>" INPUT_DETAILS_URL="https://foo"./bin
```

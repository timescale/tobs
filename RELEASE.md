# Release process and schedule

Following [Prometheus](https://github.com/prometheus/prometheus/blob/main/RELEASE.md) and [Thanos](https://github.com/thanos-io/thanos/blob/main/docs/release-process.md), this project aims for a predictable release schedule.

Release cadence of first pre-releases being cut is 4 weeks.

# How to cut a new release

> This guide is strongly based on the [Prometheus release instructions](https://github.com/prometheus/prometheus/blob/main/RELEASE.md).

## Branch management and versioning strategy

We use [Semantic Versioning](http://semver.org/).

We maintain a separate branch for each minor release, named `release-<major>.<minor>`, e.g. `release-1.1`, `release-2.0`.

The usual flow is to merge new features and changes into the `main` branch and to merge bug fixes into the latest release branch. Bug fixes are then merged into `main` from the latest release branch. The `main` branch should always contain all commits from the latest release branch.

If a bug fix got accidentally merged into `main`, cherry-pick commits have to be created in the latest release branch, which then have to be merged back into `main`. Try to avoid that situation.

Maintaining the release branches for older minor releases happens on a best effort basis.

## Update dependencies

Dependencies are updated automatically by using renovatebot. Few days before release check if there are any not merged renovate PRs and check issue called "Dependency Dashboard" to see if there are no issues with renovate.

## Update image versions

Helm charts shipped with tobs lock image versions in `chart/values.yaml`. Before doing release it is important to check and update those image versions.

## Prepare your release

For a new major or minor release, work from the `main` branch. For a patch release, work in the branch of the minor release you want to patch (e.g. `release-0.11` if you're releasing `v0.11.1`).

Update version information in multiple places:
- `version` and `appVersion` in `chart/Chart.yml`

After those steps, create a release commit with: `git commit -a -m "Prepare for the X.Y.Z release"` and create a PR for the changes to be reviewed.

## Publish the new release

For new minor and major releases, create the `release-<major>.<minor>` branch starting at the PR merge commit.
Push the branch to the remote repository with

```
git push origin release-<major>.<minor>
```

From now on, all work happens on the `release-<major>.<minor>` branch.

Tag the new release with a tag named `v<major>.<minor>.<patch>`, e.g. `v2.1.3`. Note the `v` prefix. You can do the tagging on the commandline:

```bash
tag="v1.2.3"
git tag -s "${tag}" -m "${tag}"
git push origin "${tag}"
```

Signed tag with a GPG key is appreciated, but in case you can't add a GPG key to your Github account using the following [procedure](https://docs.github.com/articles/generating-a-gpg-key), you can replace the `-s` flag by `-a` flag of the `git tag` command to only annotate the tag without signing.

The `goreleaser` GitHub Action will automatically create a new draft release with the generated binaries and a changelog attached. When it is created contact PM to validate if release notes are correct and click green publish button.

For patch releases, submit a pull request to merge back the release branch into the `main` branch.

Take a breath. You're done releasing.

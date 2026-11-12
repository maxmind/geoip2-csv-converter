# Prereqs

- You must have the [GitHub CLI tool (gh)](https://cli.github.com/) installed,
  in your path, and logged into an account that can push release tags and create
  releases in the repo.
- You must have [mise](https://mise.jdx.dev/) installed. It provides the Go and
  GoReleaser versions from `mise.lock`.
- Your environment also must have `bash`, `git`, and `sed` available.

# Releasing

- Review open issues and PRs to see if anything needs to be addressed before
  release.
- Create a release branch and switch to it. `main` is protected.
- Set the release version and today's date in `CHANGELOG.md`, for example
  `## 1.5.0 (2026-11-12)`. The version must follow
  [Semantic Versioning](https://semver.org/).
  - Mention recent changes if needed.
- Commit these changes.
- Run `dev-bin/release.sh`. It checks the GoReleaser config, pushes the branch,
  and pushes an annotated tag with the notes from `CHANGELOG.md`. Then it runs
  GoReleaser on your machine, which builds the binaries and creates the GitHub
  release with the notes. GoReleaser uses `GITHUB_TOKEN` if it is set, or else
  the token from `gh auth token`.
- Verify the release on the GitHub Releases page.
- Make a PR and get it merged.

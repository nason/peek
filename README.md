# peek

`peek` is a CLI tool that lets you deploy your static frontends to FeaturePeek. 

[FeaturePeek](https://featurepeek.com) creates **supercharged deployment previews** of your frontend that you can share (with friends, coworkers, Twitter followers — anyone!) to quickly get feedback on your implementation in progress. **A drawer overlay is added on top of your site** that makes it easy for your reviewers to take screenshots with annotations, capture screen recordings, leave comments, create tickets on popular bug-tracking platforms, and more. **You get this functionality just by deploying to FeaturePeek** — no dependencies or code changes needed.

**To get started, install from homebrew:**

```bash
brew install featurepeek/tap/peek
```

## Setup

1. `peek login` – this will create a FeaturePeek account for you if you don't have one already, and authenticate you in your CLI.
1. `peek init` – this generates a configuration file that the CLI uses.

## Usage

The typical usage flow looks like this:

1. **Commit and push your changes.** You can be on any branch.
2. **Run your build command.** Since you just committed and pushed your changes, your deployment will be tied to a hash in your git history, making it easy to see the source that generated your build.
3. **Run `peek`**. Your deployment preview will be ready after a few moments.

That's all there is to it! After your assets are packaged and uploaded, a shareable URL will be returned.

You can send this URL to anyone to get their feedback on your implementation. They won't need a FeaturePeek account to view your deployment, but they will need to create one to leave comments or file issues in the FeaturePeek drawer overlay. If you'd like your URLs to be private, subscribe to [FeaturePeek Teams](https://featurepeek.com/pricing).

## Upgrading

We periodically release new versions of this tool. To upgrade to the latest version available, run `brew upgrade peek`.

## Issues, feedback, and feature requests

Run into trouble? Have a feature request? Want to contribute? Leave any questions or ideas you may have on the [GitHub Issues](https://github.com/featurepeek/peek/issues) page.

## FeaturePeek Teams

FeaturePeek Indie is great for sharing single commits on personal projects. For company projects, you'll want to use [FeaturePeek Teams:](https://featurepeek.com/pricing)

- **Enables private deployments** that only your team can access.
- **Runs in your Continuous Integration pipeline** for automatic deployment previews on every pull request.
- **Works with frontends containerized in Docker** in addition to static frontends.

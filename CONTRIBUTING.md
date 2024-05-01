# Contribute to Developer Community Projects

Have you created something using 1Password's Developer Tools? We'd love to hear about it. Open a Pull Request to have your project listed on the [1Password Developer Portal community projects](https://developer.1password.com/community/) page.

## What can be submitted?

We're looking for **repositories** for tools, apps, and other integrations, **articles** or blog posts, or **videos** to help the community.

If the project is a repository, it should be open source. If it's an article or video, it should be public - no unlisted or paywalled content.

ℹ️ All submissions are manually reviewed by our team, and we may reject submissions for projects with dubious security practices, that appear to be inflammatory or objectionable, or overwhelmingly don't feel like a valuable contribution to the community.

## Steps to submit a project

📝 Open a [new Pull Request](https://github.com/1Password/developer-community-projects/compare) using the [default template](https://github.com/1Password/developer-community-projects/blob/main/.github/pull_request_template.md) provided.

Your PR needs to include an update to the `projects.json` file to add a new object. The new object must adhere to the following criteria:

- `category` (string) is required, and must be "article", "repo", or "video"
- `id` (string) is required, must be unique, and must only be made up of alphanumeric characters and dashes
- `title` (string) is required
- `author` (string) is required
- `url` (string) is required, must be a URL, and the URL must not redirect
- `description` (string) is optional
- `date` (string) is required, and must be in the format "YYYY-MM-DD"
- `tags` ([]string) is required

Plaintext strings should not contain URLs, emojis, angle brackets, or any non-printing characters. Please ensure changes match existing code formatting.

Finally, all commits must be signed. Not signing your commits yet? [We make this easy.](https://developer.1password.com/docs/ssh/git-commit-signing)

💬 We're also interested in helping boost projects we love. Let us know in your submission if you're interested in hearing from one of our Developer Advocates!

## Other contributions to this repository

Thanks for your interest in helping improve this repository! While we aren't actively looking for contributions of this nature, we'll still review Pull Requests.

Please ensure that you adhere to existing code styles and practices, and that your commits are signed.

## Code of Conduct

Reminder: all submissions and interactions are subject to 1Password's [Code of Conduct](https://developer.1password.com/code-of-conduct/) in order to ensure a welcoming and safe space for all.

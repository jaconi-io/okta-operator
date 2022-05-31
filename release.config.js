module.exports = {
  "branches": ["main"],
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/exec",
      {
        "prepareCmd": "make -s build-helm IMG=ghcr.io/${process.env.GITHUB_REPO}:v${nextRelease.version} VERSION=${nextRelease.version}"
      }
    ],
    [
      "@semantic-release/git",
      {
        "assets": [
          "helm"
        ],
        "message": "chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}"
      }
    ],
    "@semantic-release/github"
  ]
}
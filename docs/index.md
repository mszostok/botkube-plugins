---
sidebar_position: 1
---

# Botkube Plugins

## Why?

Botkube is a platform that gives you fast and simple access to your clusters straight from your communication platform, such as Slack, Discord, Microsoft Teams and Mattermost.

It does that by sending you notifications and allowing you to run commands. However, sometimes you may find that you need more functionality or customization than it provides. **That's where BotKube plugins come in.**

Botkube plugins allow you to extend the capabilities of BotKube and customize it to meet your specific needs.

This repository provides some useful extensions that may help you on a daily basis.

## Usage

To use plugins from this repository, configure Botkube with:

```yaml
plugins:
  repositories:
    mszostok:
      url: https://github.com/mszostok/botkube-plugins/releases/download/v1.2.0/plugins-index.yaml
```

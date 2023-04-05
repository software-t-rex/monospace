# monospace
![monospace logo](https://raw.githubusercontent.com/software-t-rex/monospace/main/docs/assets/logo-darkbg.png)

## Features
- handle multiple repositories inside a monorepo like file structure
- can run tasks against multiple projects at once concurrently (must be defined in monospace.yml)
- handle task dependencies (must be defined in monospace.yml)
- execute arbitrary commands on multiple projects at once
- display a graph representation of the task execution planning
- clone all the repositories of a monospace with a single command
- get status of all your projects with a single command
- support alias name for your projects
- Easily externalize an internal project to its own repository while preserving the same file structure
- import existing repositories in your monospace without changing anything about them and start using them in your monospace right away

## monospace != monorepository
Nowadays monorepositories are common in techs company to manage internal  dependencies through the dev team.

They came with benefits, some of which are:
- Better communication between developers teams
- Ease on-boarding of new developers in the organization by allowing them to get all source code in a single command
- Ability to launch tasks on multiple projects at once (lint, test, build ...)
- Most advanced tools allow for caching command results, implying a gain of time when working locally, and with external caching services some offers time and cost reduction on CI too.

But they also came with some drawbacks:
- Open sourcing part of a private monorepository is not simple, you end up with either an external dependency that bring us back to multi-repository paradigm. Or introducing some more tools to export part of the monorepository to an external repository, which will make it more cumbersome when it comes to accept external Pull requests
- Multi-repositories allow for fine grained access control, whether it's not impossible to achieve in monorepositories, it's not an easy task either (some tools have this feature built-in but then you rely on a new access control system to keep in sync).

**So here come monospace**
monospace aim to provide you with the best possible subset of this too worlds:
- Easy on-boarding allowing to get all the necessary source code in a single command line
- Same benefits as monorepo for cross team communication
- Ability to launch tasks through multiple projects at once
- Fine grained access control with tools you already have in place (git repository access)
- Easy separation between private and public projects without additional cost for accepting external contributions

## Why this name ?
Nothing really fancy here, but a monorepository is compound of workspaces, so contracting this led to monospace, which is the word in french for family minivan vehicles. I found it being sufficiently expressive about the nature of the project and funny enough for us to like it, so yeah "monospace" :)

## How does it work ?
The main idea, while not new is inspired by [meta](https://github.com/mateodelnorte/meta)-repository a javascript project which concept is explained [here](https://patrickleet.medium.com/mono-repo-or-multi-repo-why-choose-one-when-you-can-have-both-e9c77bd0c668). Monospace main difference is it's written in go and should work with different package managers out of the box. By speaking of package manager, you will understand that monospace as many monorepository tools is mainly targeted for javascript/typescript projects. But there's no restriction on using it for other kind of projects, as monospace is a go project that is developed under a monospace.
Only some commands maybe specifically made for js ecosystem.

So how does it work ?
In a monospace you can have many kind of projects like applications, libraries, ...
But in monospace terms there is three kind of projects:
"internal", "external" and "local":
- internal: are projects which are embedded in the monospace repository and share the same history.
- external: are projects that has their own repositories inside the monospace repository. They are gitignored by the monospace repository, and so they manage their own history.
- local: this is exactly the same as external projects but without a configured remote repository, and so they are not published anywhere, but their name is reserved in the global monospace so the name won't be taken by another developer accidentally. This case can be used by a new library project you have not shared yet with the rest of the organization.

## How to get started / Documentation
Go to the [documentation directory](https://github.com/software-t-rex/monospace/blob/main/docs/monospace/index.md) and follow instructions there.

## Contact us
Have questions? Want to talk about monospace? Want announcements for new versions, You can join our discord server at https://discord.gg/WHdZkqh7gA

## Contributing
monospace is open source software under the MIT license, and by contributing to it you accepts to release your code under the same License. Contributions are always welcomed and will be reviewed in the shortest amount of time as possible. If you decide to contribute, please make small organized commits that address one thing at a time, it will make it easier for us to review and accept your contribution.

## Funding / Sponsorship
This project is free software, but to live it needs time, and to get time you needs money. So if this project is of any help to you or your company, and/or you want to make it evolve quicker, you can [become sponsors to the project](https://github.com/sponsors/malko).

Donation over 1000€ will allow you or your company to appears on this page as sponsors of the project, in such case contact us at contact.trex.software@gmail.com with the receipt of your donation.

# monospace

## monospace != monorepository
Nowadays monorepositories are common in techs company to manage internal  dependencies through the dev team.

They came with benefits, some of which are:
- Better communication between developers teams
- Ease onboarding of new developpers in the organization by allowing them to get all source code in a single command
- Ability to launch tasks on multiple projects at once (lint, test, build ...)
- Most advanced tools allow for caching command results, implying a gain of time when working locally, and with external caching services some offers time and cost reduction on CI too.

But they also came with some drawbacks:
- Open sourcing part of a private monorepository is not simple, you end up with either an external dependecy that bring us back to multi-repository paradigm. Or introducing some more tools to export part of the monorepository to an external repository, which will make it more cumbersome when it comes to accept external Pull requests
- Multi-repositories allow for fine grained access control, whether it's not impossible to achieve in monorepositories, it's not an easy task either (some tools have this feature built-in but then you rely on a new access control system to keep in sync).

**So here come monospace**
monospace aim to provide you with the best possible subset of this too worlds:
- Easy onboarding allowing to get all the necessary source code in a single command line
- Same benefits as monorepo for cross team communication
- Ability to launch tasks through multiple projects at once
- Fine grained access controll with tools you already have in place (git repository access)
- Easy separation between private and public projects without additional cost for accepting external contributions

## Why this name ?
Nothing really fancy here, but a monorepository is compound of workspaces, so contracting this led to monospace, which is the word in french for family minivan vehicules. I found it being sufficiently expressive about the nature of the project and funny enough for me to like it, so yeah "monospace" :)

## How does it work ?
The idea, while not new is inspired by meta-repository which is a javascript project, monospace main difference is it's written in go and should work with different package managers out of the box. By speaking of package manager, you will understand that monospace as many monorepository tools is mainly targeted for javascript/typescript projects. But there's no restriction on using it for other kind of projects, as monospace is a go project that is developped under a monospace.
Only some commands maybe specificly made for js ecosystem.

So how does it work ?
In a monospace you can have many kind of projects like applications, libraries, ...
But in monospace terms there is three kind of projects:
"internal", "external" and "local":
- internal: are projects which are embedded in the monospace repository and share the same history.
- external: are projects that has their own repositories inside the monospace repository. They are gitignored by the monospace repository, and so they manage their own history.
- local: this is exactly the same as externdal projects but without a configured remote repository, and so they are not published anywhere, but their name is reserved in the global monospace so the name won't be taken by another developper accidentally. This case can be used by a new library project you have not shared yet with the rest of the organisation.

## requirements
- git command should be available in your Path
- js package manager defined in your config should also be available in your path
	(can be omit if you don't plan to work on javascript projects)
- go should be available in your path if you intend to work on go projects

## Some Default opinionated choices:
When initializing a new monospace it will declare some workspaces to your package manager:
- apps/* for applications
- packages/* for libraries

Default to add the following to monospace gitignore file (this will also be applied to local project created with monospace):
- node_modules
- .vscode
- .env
- dist
- coverage

Default package manager is pnpm 7, for now this is the only one tested but it should work correctly with yarn or npm, don't hesitate to report any issues with this pacakge managers, they should be first citizen too.

monospace .npmrc will contains the following default setttings
- auto-install-peers=true
- resolve-peers-from-workspace-root=true

If you think that this is not a good default feel free to contact me and explain why you think we should use other defaults. I'm always prone to change my mind about such decisisons when there's good reasons to do so.

## Contributing
monospace is open source software under the MIT license, and by contributing to it you accepts to release your code under the same License. Contributions are always welcomed and will be reviewed in the shortest amount of time as possible. If you decide to contribute, please make small organized commits that adress one thing at a time, it will make it easier for me to review and accept your contribution.

## Funding
This project is free software, but to live it needs time, and to get time you needs money. So if this project is of any help to you and/or you want to make it evolve quicker, you can make a donation or support the project through this button.
<div id="donate-button-container">
<div id="donate-button"></div>
<script src="https://www.paypalobjects.com/donate/sdk/donate-sdk.js" charset="UTF-8"></script>
<script>
PayPal.Donation.Button({
env:'production',
hosted_button_id:'SRHXHER2G48CA',
image: {
src:'https://www.paypalobjects.com/en_US/i/btn/btn_donate_LG.gif',
alt:'Donate with PayPal button',
title:'PayPal - The safer, easier way to pay online!',
}
}).render('#donate-button');
</script>
</div>
Donation over 1000€ will allow you to appears on this page as sponsors of the project, in such case contact us at contact.trex.software@gmail.com with the receipt of your donation.
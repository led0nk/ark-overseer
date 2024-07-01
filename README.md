<div align="center">

  <img src="https://github.com/led0nk/ark-overseer/assets/10290002/3b420707-4385-4ff1-a4fb-cdc42e1e75a1" width=500>
 

# Ark-Overseer - observation tool

[Installation](#installation)
•
[Messaging](#messaging)
•
[Contribution](#contribution)

</div>




## Summary

Ark-Overseer is a handmade application to observe as many Ark-Servers as you want to.
It is capable of tracking players via their `Steam-Name`. Since it's common case to use
the `Steam-Name` `123` it might not be the best application for official servers (a `Steam-ID`-implementation is planned for later releases).
You can simply add the servers you'd wish to track via the web-interface:

![swappy-20240603-135719](https://github.com/led0nk/ark-overseer/assets/10290002/afbf8d2e-aaa7-421d-9fb1-7ac34e38cb60)



The messaging feature can be configured through the `Settings`-tab in the navigation-bar.
See more -> [Messaging](#messaging)

The tracked players can be configured via their `Steam-Name` on the `Blacklist`-tab in the navigation bar:

![swappy-20240603-135636](https://github.com/led0nk/ark-overseer/assets/10290002/40589b09-7e23-44f6-9b5a-5baace7e0337)

## Installation

### via rpm

The Ark-Overseer is also available as an rpm-package, which makes it easy to install.

##### 1.) First you have to enable the repository:

```sh 
dnf copr enable led0nk/ark-clusterinfo
```

##### 2.) Then you're able to simply install it via the dnf-pkg-manager:

```sh 
dnf install ark-overseer
```

##### 3.) After finishing the installation you can simply run it via cli-command:

```sh 
ark-overseer
```

##### Note for systemd-service:

When installing the application via rpm, you will get a systemd service installed
and enabled, which ensures to launch the application on system startup.

Therefore the directories for the database and blacklist are expected like the following:

| component | directory |
| -------------- | --------------- |
| database | /etc/ark-overseer/ |
| blacklist | /etc/ark-overseer |



### via Docker

The most simple way of installation is to just run the application in a container.
E.g. with a container-service like Docker:

```sh 
docker run -it \
    -p 8080:8080 \
    --rm ghcr.io/led0nk/ark-overseer:latest
```

## Messaging 

### Setup for Discord-Bot

There are some steps to follow through for getting a working Discord-Bot.

#### 1.) Create an app/bot

You should just follow the first step `Creating an app` in this guide here:
[Discord-Bot](https://discord.com/developers/docs/quick-start/getting-started)
**Make sure to write down the `token` of your bot.**

#### 2.) Permissions

Once you've added the Bot to your Discord-Server, you should configure it's permissions.
It should at least be able to:
  - view channels
  - send messages

On top of that you have to verify that the `notification-channel` grants the 
permissions to:
  - view channel
  - send messages

#### 3.) Developer Options

Now you have to enable the developer options in discord, because you need to
get the `channel-ID`, where your bot will send notifications.

Therefore you go into the user-settings:

![swappy-20240603-151145](https://github.com/led0nk/ark-overseer/assets/10290002/46f3b055-5483-4d0f-aaba-4bf2779f5857)



After enabling the dev-mode, you can now copy the `channel-ID` to set up the notifications.


![swappy-20240603-151204](https://github.com/led0nk/ark-overseer/assets/10290002/557c2963-f5ae-4144-b1ea-b1edc993f925)


Now you're able to fill in the Notification-Settings with the `token`and the `channel-ID` and run your discord-bot.


![swappy-20240603-135404](https://github.com/led0nk/ark-overseer/assets/10290002/3f35ec51-ee70-4188-85f8-36cb6ebc383f)


## Contribution

If you're interested in improving the code quality or enhancing the features of
the Ark-Overseer, you're very welcome to contribute to this repository.

To contribute, simply fork the repository, make your changes, submit a pull request
and add some information on the changes. Your contributions will be reviewed as
soon as it is possible.

Don't hesitate to open issues, when being confronted with the applications bugs.


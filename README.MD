# Warp Bot
A bot to give you wireguard profiles that use [warp](https://1.1.1.1) as their server.

**Note that this project does not have anything to do with real cloudflare and this is not an official bot for warp.**
## Downloading And Installing
At first install golang. Then clone this repository with `git clone https://github.com/HirbodBehnam/Warp-Bot`.

If you want, you can edit main.go and change `const WarpPlusEnabled = false` to `const WarpPlusEnabled = true` in order to enable the warp+ quota adder.

Now execute these commands in order to build the bot
```bash
cd Warp-Bot
go get github.com/HirbodBehnam/EasyX25519 github.com/allegro/bigcache github.com/go-telegram-bot-api/telegram-bot-api
go build main.go profile.go
```

Then you run the executable file and pass your token as the first argument.

## Warp Plus
This bot has an optional feature that you can use in order to allow your users add 1GB of warp plus to their account each 24 hours.

By default, this option is disabled. But it can be re-enabled at compile time by changing `const WarpPlusEnabled = false` to `const WarpPlusEnabled = true` in main.go file.

The problem is that that your IP might get banned due to flooding.

## Non-Bot version
If you want you can use [wgcf](https://github.com/ViRb3/wgcf/) in order to get wireguard config with your own PC. Also, I just wanna thanks him because some pars of my code is scrapped from that project.
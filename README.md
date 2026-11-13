# Brave Portable

Portable builds of **Brave Stable**, **Brave Beta** and **Brave Nightly** for Windows.
Run them from a USB drive or any folder. Your profile stays next to the exe.

## Downloads

Grab the exe you want from the [Releases](../../releases) page:

| Channel | File |
|---------|------|
| Stable  | `Brave-Stable-Portable.exe` |
| Beta    | `Brave-Beta-Portable.exe` |
| Nightly | `Brave-Nightly-Portable.exe` |

## Install

1. Put the exe in its own empty folder (e.g. on your USB drive).
2. Run it. A setup window opens with the Brave logo for that channel.
3. Click **Install**. The latest build is downloaded into the `app` folder.
4. Click **Launch**. Next time you run the exe, Brave starts directly.

## Folder layout

```
BravePortable/
├── Brave-Beta-Portable.exe
├── app/     Brave itself
└── data/    your profile (bookmarks, extensions, settings)
```

## Update (manual)

1. Close Brave.
2. Delete the **`app`** folder. Do **not** delete `data`.
3. Run the exe again and click **Install**.

Your profile in `data` is kept.

## Notes

- Windows 10/11, 64-bit only.
- Use a separate folder for each channel so profiles don't mix.
- Brave is downloaded from the official [brave/brave-browser](https://github.com/brave/brave-browser) releases.
- This is an unofficial project and is not affiliated with Brave Software.

## Build

Push a tag (`v*`). GitHub Actions builds the three exes and attaches them to the release.

```
git tag v1.0.0
git push --tags
```

Logos live in `assets/` (`stable.png`, `beta.png`, `nightly.png`). They are embedded in the
setup window and used as the exe icon.

## Credits

Logos are Brave's official product icons, taken from
[brave/brave-core](https://github.com/brave/brave-core/tree/master/app/theme/brave/linux).
Brave and the Brave logos are trademarks of Brave Software.

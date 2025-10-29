<H1 align="center">
  Yappers of Linux
</H1>

<p align="center">
  <img src="other/assets/yappers-of-linux-banner.png"/>
</p>

<p align="center">Voice typing for Linux that actually works</p>

<p align="center">
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/stargazers">
    <img src="https://img.shields.io/github/stars/DeprecatedLuar/yappers-of-linux?style=for-the-badge&logo=github&color=1f6feb&logoColor=white&labelColor=black"/>
  </a>
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/releases">
    <img src="https://img.shields.io/github/v/release/DeprecatedLuar/yappers-of-linux?style=for-the-badge&logo=go&color=00ADD8&logoColor=white&labelColor=black"/>
  </a>
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/DeprecatedLuar/yappers-of-linux?style=for-the-badge&color=green&labelColor=black"/>
  </a>
</p>

---

## Somehow I had to build this

Voice typing on Linux either doesn't work or was made in the past century. How android beats Linux on that for the past 10 years? I can't let that slide.

---

## How it works

Uses faster-whisper (optimized Whisper implementation) with a circular pre-recording buffer and WebRTC VAD for speech detection. The whole engine is embedded in a single Go binary that self-extracts and manages its own Python environment using hashed files to check dependencies and update them as needed.

## The cool features you've never seen before

- **Portable binary** - Literally one file, zero setup, runs anywhere (well, at least Linux)
- **Performance modes** - Faster or more accurate modes based on your hardware
- **One toggle** - `yap toggle` pauses, resumes, or starts (control via cli)
- **State server** - TCP server for status bar/widget integrations
- **Actually private** - Literally Whisper (Wow!)

---

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/DeprecatedLuar/yappers-of-linux/main/install.sh | bash
```

<details>
<summary>Other Install Methods</summary>

<br>

**Manual Install**
```bash
# Download binary from releases
wget https://github.com/DeprecatedLuar/yappers-of-linux/releases/latest/download/yap
chmod +x yap
sudo mv yap /usr/local/bin/
yap start  # Auto-installs everything
```

**Build From Source**
```bash
git clone https://github.com/DeprecatedLuar/yappers-of-linux.git
cd yappers-of-linux
go build -o yap cmd/main.go
./yap start
```

**System Requirements**:
- `python3` (3.10+)
- `portaudio19-dev` (for mic access)
- `ydotool` + `ydotoold` (Wayland) OR `xdotool` (X11)

</details>

First run takes ~2 minutes to download and set up everything. After that, it's instant.

---

## Usage

```bash
yap start                     # Start listening
yap start --model small       # Use better model
yap start --fast              # Faster but less accurate
yap toggle                    # Pause/resume/start
yap stop                      # Stop
```

<details>
<summary>Available Models</summary>

<br>

Models auto-download on first use:

| Model  | Size   | Speed      | Accuracy |
|--------|--------|------------|----------|
| tiny   | ~75MB  | Fastest    | Basic    |
| base   | ~150MB | Fast       | Good     |
| small  | ~500MB | Balanced   | Better   |
| medium | ~1.5GB | Slow       | Great    |
| large  | ~3GB   | Slowest    | Best     |

</details>

<details>
<summary>More stuff you can do</summary>

<br>

```bash
yap start --device cuda       # Use GPU instead
yap start --language es       # Spanish (or any other language)
yap start --tcp               # Enable state server on port 12322
yap start --no-typing         # Just prints to terminal, doesn't type
yap models                    # See what models you have
yap config                    # Open config in your editor
```

</details>

<details>
<summary>Configuration</summary>

<br>

Config file lives at `~/.config/yappers-of-linux/config.toml` and gets created on first run.

```toml
notifications = "start,urgent"   # When to notify you
model = "tiny"                   # Which model to use
device = "cpu"                   # cpu or cuda
language = "en"                  # What language you're speaking
fast_mode = false                # Trade accuracy for speed
```

Run `yap help config` if you want all the details.

</details>

<details>
<summary>How It Works (for the nerds)</summary>

<br>

**The trick**: Keeps a 1.5-second audio buffer running constantly. When voice detection kicks in, it grabs that buffer first so your opening words aren't lost.

**What happens**:
1. Always recording to a circular buffer
2. Voice detected → saves buffer + keeps going
3. You stop talking for 0.8s → triggers transcription
4. Whisper does its thing → types the text

**Why it just works**: The binary has the entire Python engine baked in. First run unpacks it, sets up a venv, installs what it needs. Updates handle dependency changes on their own. You never mess with Python directly.

</details>

<details>
<summary>State Monitoring (if you're into that)</summary>

<br>

Want to plug this into your status bar or a desktop widget?

```bash
yap start --tcp        # Starts on port 12322
nc 127.0.0.1 12322     # Test it out
```

Spits out JSON with the current state. Inspired by [Kanata's TCP port](https://github.com/jtroo/kanata).

</details>

---

<p align="center">
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/issues">
    <img src="https://img.shields.io/badge/Found%20a%20bug%3F-Report%20it!-red?style=for-the-badge&logo=github&logoColor=white&labelColor=black"/>
  </a>
</p>

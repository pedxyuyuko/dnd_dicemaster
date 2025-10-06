# DnD DM - The Dice Master
[中文版](README_CN.md) | [English](README.md)

Telegram @dndicem_bot
Roll the dice for you

# Install

## Docker

Image: ``uuz233/telegram_dicemaster_bot``

### Compose
```yaml
services:
  tg_dicemaster_bot:
    image: uuz233/telegram_dicemaster_bot
    env_file: config.env
    stop_grace_period: "2"
```

## Local Build

1. Clone the repository:
   ```bash
   git clone https://github.com/pedxyuyuko/dnd_dicemaster.git
   cd dnd_dicemaster
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Copy the example environment file:
   ```bash
   cp example.env .env
   ```

4. Edit `.env` with your bot token and optional settings.

5. Build:
   ```bash
   go build
   ```

6. Run the binary:
   ```bash
   ./dnd_dicemaster
   ```

# Features

- **Dice Rolling**: Supports standard D&D dice notation (e.g., 1d20, 2d6+3)
- **Advantage/Disadvantage**: Use 'A' for advantage, 'D' for disadvantage
- **Modifiers**: Add or subtract values (e.g., 1d20+5-2)
- **Cryptographically Secure Randomness**: Uses Ethereum block hashes for true randomness
- **Inline Telegram Queries**: Works directly in Telegram chats
- **Limits**: Dice count ≤ 1000, faces ≤ 1000

# Usage

The bot responds to inline queries in Telegram. Start typing `@dndicem_bot` followed by your dice command.

## Basic Dice Roll
- `1d20` - Roll a single 20-sided die
- `2d6` - Roll two 6-sided dice

## With Modifiers
- `1d20+5` - Roll d20 and add 5
- `1d20+3-2` - Roll d20, add 3, subtract 2

## Advantage/Disadvantage
- `A 1d20` - Roll with advantage (take higher of two rolls)
- `D 1d20` - Roll with disadvantage (take lower of two rolls)

## Skill Checks
- `1d20>15` - Roll d20, check if ≥ 15
- `A 1d20+2>12` - Advantage roll with modifier and check

## Named Checks
- `Strength A 1d20+3>10` - Named attribute check with advantage

# Configuration

Create a `.env` file based on `example.env`:

- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token from @BotFather
- `TELEGRAM_API`: Telegram API endpoint (default: https://api.telegram.org)
- `ETH_RPC_URL`: Ethereum RPC URL for randomness (default: https://eth.llamarpc.com)

# License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

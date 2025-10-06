# DnD DM - The Dice Master
Telegram @dndicem_bot  
扔骰子！大成功！

# 安装

## Docker

镜像: ``uuz233/telegram_dicemaster_bot``

## 本地构建

1. 克隆仓库:
   ```bash
   git clone https://github.com/pedxyuyuko/dnd_dicemaster.git
   cd dnd_dicemaster
   ```

2. 安装依赖:
   ```bash
   go mod download
   ```

3. 复制示例环境文件:
   ```bash
   cp example.env .env
   ```

4. 编辑 `.env` 文件，填入您的机器人令牌和可选设置。

5. 构建:
   ```bash
   go build
   ```

6. 运行二进制文件:
   ```bash
   ./dnd_dicemaster
   ```

# 功能特性

- **掷骰**: 支持标准的 D&D 骰子记法 (例如: 1d20, 2d6+3)
- **优势/劣势**: 使用 'A' 表示优势, 'D' 表示劣势
- **修正值**: 添加或减去数值 (例如: 1d20+5-2)
- **技能检定**: 与阈值比较结果 (例如: 1d20>15)
- **加密安全的随机性**: 使用以太坊区块哈希确保真正的随机性
- **Telegram 内联查询**: 直接在 Telegram 聊天中工作
- **限制**: 骰子数量 ≤ 1000, 面数 ≤ 1000

# 使用方法

机器人响应 Telegram 中的内联查询。开始输入 `@dndicem_bot` 后跟您的骰子命令。

## 基本掷骰
- `1d20` - 掷一个 20 面骰
- `2d6` - 掷两个 6 面骰

## 带修正值
- `1d20+5` - 掷 d20 并加 5
- `1d20+3-2` - 掷 d20，加 3，减 2

## 优势/劣势
- `A 1d20` - 优势掷骰 (取两次掷骰中的较高值)
- `D 1d20` - 劣势掷骰 (取两次掷骰中的较低值)

## 技能检定
- `1d20>15` - 掷 d20，检查是否 ≥ 15
- `A 1d20+2>12` - 优势掷骰带修正值并检定

## 命名检定
- `Strength A 1d20+3>10` - 命名属性检定带优势

# 配置

基于 `example.env` 创建 `.env` 文件:

- `TELEGRAM_BOT_TOKEN`: 您的 Telegram 机器人令牌，从 @BotFather 获取
- `TELEGRAM_API`: Telegram API 端点 (默认: https://api.telegram.org)
- `ETH_RPC_URL`: 以太坊 RPC URL 用于随机性 (默认: https://eth.llamarpc.com)

# 许可证

本项目采用 MIT 许可证 - 详情请见 [LICENSE](LICENSE) 文件。

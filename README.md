# AI Companion CLI (Go Edition)

![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)
![TUI](https://img.shields.io/badge/TUI-Charmbracelet-FF69B4)

一个专为终端（Terminal）打造的高沉浸、高性能 AI 专属陪伴程序。本项目已从早期的 Python POC（概念验证）版本，全面深度重写为 **Go (Golang)**，带来极致的加载速度、毫秒级的 UI 响应，以及彻底免疫源码反编译的安全性保护。

## ✨ 核心特性

- **🚀 极致性能与零依赖**：基于 Go 语言编写，编译后为单一的机器码二进制文件。无需像 Python 那样痛苦地配置依赖，**下载即玩，双击即用**，无论是小白用户还是资深极客都能零门槛上手。
- **💄 惊艳的终端美学 (Charmbracelet 生态)**：抛弃原有的粗糙交互，采用 `Bubbletea` 和 `Lipgloss` 重新打造。流畅的滚动面板、高定制的 CSS 级终端气泡、以及优美的交互按键，提供极致的终端体验。
- **🔒 核心 Prompt 级保护**：通过 Go 预编译为机器指令，避免了传统脚本语言（Python 等）打包后极易被工具反编译（如 PyInstxtractor）而泄露核心 Prompt 资产和业务逻辑的致命弱点。
- **🧠 进阶的 AI 记忆与亲密度系统**：通过本地轻量级 SQLite 数据库（基于 GORM 驱动）原生地存储角色配置（包含定制的外貌、性格MBTI、爱好等）和深度的中长期对话记忆，带来真实的“养成”陪伴感。
- **⚡️ 完美的流式响应 (Streaming)**：对接最新的 OpenAI 协议接口，在您的终端中打字般逐字渲染，拒绝等待，体验自然对话的心流。

> [!IMPORTANT]
> **API 限制说明**：本项目目前**仅支持 OpenAI 官方 API 模式**（及完全兼容 OpenAI SDK 格式的第三方大模型 API，如 DeepSeek、通义千问等，只需替换 BaseURL）。你需要拥有一个有效的 API Key 才能驱动 AI 伴侣。

---

## 📦 如何使用 (对普通用户)

你**完全不需要了解任何代码**即可使用！

### 第一步：配置你的 API Key
程序需要读取大语言模型的 API 密钥才能工作。
打开你的终端，设置环境变量（你也可以在程序同级目录下创建一个 `.env` 文件并写入 `OPENAI_API_KEY=sk-...`）：

**Mac/Linux:**
```bash
export OPENAI_API_KEY="sk-你的真实APIKey"
```
**Windows (PowerShell):**
```powershell
$env:OPENAI_API_KEY="sk-你的真实APIKey"
```

### 第二步：运行程序
1. 获取发布包的二进制文件。
2. 运行对应你操作系统的文件：
   - **Windows 用户**: 双击 `ai-companion-windows.exe` 运行。
   - **Mac 用户**: 在终端中进入所在目录，执行 `./ai-companion-macos`。
   - **Linux 用户**: 赋予执行权限后运行（例如 `chmod +x ai-companion-linux && ./ai-companion-linux`）。
3. 按照屏幕上优美的 UI 提示，填入你的大模型 API 密钥。
4. 开始创建你的专属 AI 伴侣体验！

---

## 🛠️ 如何开发 (对极客和开发者)

### 环境要求
- [Go 1.22+](https://go.dev/dl/)

### 本地编译运行

进入项目根目录：
```bash
cd ai-companion-cli-go
```

自动安装所有依赖包（TUI、GORM、OpenAI SDK 等）：
```bash
go mod tidy
```

直接运行项目（开发模式）：
```bash
go run cmd/cli/main.go
```

编译为你自己电脑操作系统的独立程序：
```bash
go build -o ai-companion cmd/cli/main.go
```

### 交叉编译 (打包出Windows/Mac/Linux版本分发)

由于 Go 无敌的交叉编译能力，你可以在任何电脑上一键打出所有平台的包！

**打出 Windows 的 EXE：**
```bash
GOOS=windows GOARCH=amd64 go build -o ai-companion.exe cmd/cli/main.go
```

**打出 Mac (苹果芯片 M1/M2/M3) 的程序：**
```bash
GOOS=darwin GOARCH=arm64 go build -o ai-companion-mac-arm cmd/cli/main.go
```

**打出 Linux 的程序：**
```bash
GOOS=linux GOARCH=amd64 go build -o ai-companion-linux cmd/cli/main.go
```

---

## 🏗 架构说明

本项目遵循清晰的模块化设计思想，非常易于二次开发：
- `cmd/cli/`：程序入口点，包含了编译的 main 函数。
- `internal/ui/`：一切跟界面相关的代码。使用了 `Bubbletea` 的 Model-Update-View 架构处理复杂的交互状态机。
- `internal/orchestrator/`：中枢大脑单元。负责串联用户输入、调用记忆、计算亲密度、然后组装 Prompt 发往后端。
- `internal/storage/`：以 SQLite + GORM 构建的持久化存储方案，包含各种数据模型映射。
- `internal/llm/`：纯粹的大模型交互封装层。

## 📝 License
本项目基于 **GNU AGPLv3 (Affero General Public License v3.0)** 协议开源。
这意味着您可以自由地使用、修改和分发本代码，但任何基于本代码修改或通过网络提供服务（如部署在服务器上供他人使用）的作品，都必须同样以 AGPLv3 协议开源其完整的源代码。

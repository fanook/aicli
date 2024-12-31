# aicli

![Build Status](https://img.shields.io/github/actions/workflow/status/fanook/aicli/release.yml)
![GitHub License](https://img.shields.io/github/license/fanook/aicli)
![Version](https://img.shields.io/github/v/release/fanook/aicli)


## 描述
一款基于 AI 的命令行工具集，旨在为开发者与程序员提供便捷、高效、有趣的命令行体验。
![2024-12-3112 51 17-ezgif com-video-to-gif-converter](https://github.com/user-attachments/assets/bcd7823f-5c5a-44fb-afb2-315a7f005f65)



## 功能特性

| 命令          | 描述                                                         | 示例命令                       |
|---------------|--------------------------------------------------------------|------------------------------|
| `chat`        | 与 AI 进行持续对话                                             | `aicli chat`                 |
| `git-cmt`     | 智能分析 Git Changes，生成更语义化、更易读的 commit message。    | `aicli git-cmt`              |
| `gen-cmd`     | 根据自然语言描述自动生成命令行语句，并附带解释说明。              | `aicli gen-cmd 查看磁盘大小` |
| `joke`        | 随机讲一个与程序员相关的笑话，缓解工作压力。                      | `aicli joke`                 |



## 安装
### 1. 使用Go安装
```shell
go install github.com/fanook/aicli@latest
```
### 2. 验证安装
```shell
aicli help 
```
### 3. 配置依赖的环境变量
```dotenv
# AICLI_OPENAI_API_KEY 是必需配置的 OpenAI API 密钥。
AICLI_OPENAI_API_KEY=sk-sccvcat-qXAMosGXIrEs-MT3FqNOkGhGsOBcZ3XtJ6O_pbgeFJ_u9uwT3szVHYcjMZOYqf2Jv8WcUVTKKmAEtkCtjjrenHbc5zESoczT3BlboLGuUbRCTCYMVp5wr15Z64c6e4ykWcmc4rAA

# 以下为可选配置，当前显示为默认值。
AICLI_OPENAI_MODEL=gpt-4o-mini
AICLI_OPENAI_API_URL=https://api.openai.com/v1/chat/completions
AICLI_GITCOMMIT_PROMPT="你是一个帮助生成 Git commit 信息的助手。请根据以下 Git 仓库的变更生成一个简洁且有意义的 Git commit 信息。请严格遵循以下格式，并且只能使用以下两种类别：\n\n[类别] 描述\n\n**可用类别：**\n- **feat**: 新功能\n- **fix**: 修复\n\n**示例：**\n[fix] 修复用户登录时的验证错误\n[feat] 添加用户个人资料页面\n\n变更内容：\n{{.Changes}}"
AICLI_GENCMD_PROMPT="你是一个帮助生成命令行指令和解释的助手, 请根据以下描述生成一个适合当前机器的命令行指令，并提供简要的解释：描述：{{.Description}} 操作系统：{{.OS}} 架构：{{.Arch}}  生成的格式举例(严格按照此格式)： CMD: free -m \n 解释: 显示当前系统内存使用情况"
AICLI_JOKE_PROMPT="你是一个讲程序员相关笑话的助手, 请生成一个与程序员相关的笑话： 生成的格式举例（严格按照此格式）： 为什么程序员总是混淆圣诞节和万圣节？因为 Oct 31 == Dec 25！ 因为在八进制中，31 等于十进制的 25。"
AICLI_CHAT_PROMPT="你是一个智能聊天助手，能够与用户进行自然流畅的对话。"
```

### 4. 开始使用
```shell
aicli chat
```

### 5. 更简洁的使用
```shell
# 添加别名（选择合适的shell配置文件，这里以bash_profile举例）
echo "alias ac='aicli chat'" >> ~/.bash_profile
# 重新加载shell配置
source ~/.bash_profile
# 验证别名
ac
```

## 贡献指南
如果您在使用或开发过程中遇到问题，欢迎在 Issues 页面提交问题或讨论。欢迎任何形式的贡献！

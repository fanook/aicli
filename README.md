# aicli

![Build Status](https://img.shields.io/github/actions/workflow/status/fanook/aicli/release.yml)
![GitHub License](https://img.shields.io/github/license/fanook/aicli)
![Version](https://img.shields.io/github/v/release/fanook/aicli)


## 描述
一款基于 AI 的命令行工具集，旨在为开发者与程序员提供便捷、高效、有趣的命令行体验。

![20250102-114813-ezgif com-video-to-gif-converter](https://github.com/user-attachments/assets/56a8b93d-b7e4-4fff-b290-d2007629200b)



## 功能特性
| 示例命令                            | 描述                               |
|-------------------------------------|----------------------------------|
| `aicli chat`                        | 与AI进行持续对话。                       |
| `aicli git-cmt`                     | 智能分析Git Changes生成Commit Message。 |
| `aicli gen-cmd 查看磁盘大小`        | 根据自然语言描述生成命令行语句。                 |
| `aicli joke`                        | 讲一个与程序员相关的笑话。                    |
| `aicli process-data`                | 批量数据处理。                          |

## 更重要的功能-批量数据处理
### 使用场景
- 针对大规模数据集进行自动化数据清洗、数据转换等操作。
- 例子：文本批量打标签
  - 有一个场景需要使用AI对一些聊天文本进行打标签，判断是否包含脏话，数据源为DB或CSV。
  - 通过aicli process-data 可以快速实现这个场景，只需要提供数据源、全局的prompt或者行自定义prompt, 即可逐行生成结果至result。

### 数据规则
- csv或db的字段确保为 id(number),content(string),prompt(string),result(string))。
- 示例csv：
```csv
id,content,prompt,result
1,你好,你好瘦{{.Content}}是否是脏话,你只需回复是或否,否
2,你好,fuck{{.Content}}是否是脏话，你只需回复是或否,是
```
- db表结构：
```sql
CREATE TABLE my_table_name (
    id INT AUTO_INCREMENT PRIMARY KEY,
    content VARCHAR(1000) NOT NULL DEFAULT '',
    prompt  VARCHAR(1000) NOT NULL DEFAULT '',
    result  VARCHAR(1000) NOT NULL DEFAULT ''
);
```
### 使用方法
```shell
# 处理csv数据
aicli process-data -s csv -f test/my_data.csv -o test/my_data_result.csv

# 处理db表数据
aicli process-data -s db --db-host 127.0.0.1 --db-port 3306 -u root -P mydbpassward --db-name my_db_name --db-table my_table_name
```

## 安装和使用
### 1. 安装
#### 方法1: 通过Go安装
如果您已经安装了 Go 开发环境，可以直接通过以下命令安装：
```shell
go install github.com/fanook/aicli@latest
```
#### 方法2: 下载预编译的二进制文件安装
- 前往 [Releases 页面](https://github.com/fanook/aicli/releases) 下载适合您操作系统的执行文件。
- 下载完成后，将文件移动到系统的 PATH 路径中，例如 /usr/local/bin。


### 2. 验证安装
```shell
aicli -v
```
### 3. 配置依赖的环境变量
```shell
export AICLI_OPENAI_API_KEY=sk-sccvcat-qXAMosGXIrEs-MT3FqNOkGhGsOBcZ3XtJ6O_pbgeFJ_u9uwT3szVHYcjMZOYqf2Jv8WcUVTKKmAEtkCtjjrenHbc5zESoczT3BlboLGuUbRCTCYMVp5wr15Z64c6e4ykWcmc4rAA
```
```dotenv
# AiProvider: AI服务提供商，目前可选值为 openai 或 deepseek, 默认为 openai
AICLI_PROVIDER=deepseek

# Openai: 如果您选择使用 OpenAI 作为AI服务提供商，AICLI_OPENAI_API_KEY为必填项
AICLI_OPENAI_API_KEY=sk-sccvcat-qXAMosGXIrEs-MT3FqNOkGhGsOBcZ3XtJ6O_pbgeFJ_u9uwT3szVHYcjMZOYqf2Jv8WcUVTKKmAEtkCtjjrenHbc5zESoczT3BlboLGuUbRCTCYMVp5wr15Z64c6e4ykWcmc4rAA
AICLI_OPENAI_MODEL=gpt-4o
AICLI_OPENAI_API_URL=https://api.openai.com/v1/chat/completions

# Deepseek: 如果您选择使用 Deepseek 作为AI服务提供商，AICLI_DEEPSEEK_API_KEY为必填项
AICLI_DEEPSEEK_API_KEY=sk-f122522c61892d95a0e5f4174e21563f
AICLI_DEEPSEEK_MODEL=deepseek-chat
AICLI_DEEPSEEK_API_URL=https://api.deepseek.com/chat/completions

# Prompts: cmd的预设prompt，您也可以自定义或在cmd中以prompt参数传递。
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
# 更多别名
alias acl='aicli'
alias agc='aicli git-cmt'
alias ac='aicli chat'
alias acd='aicli gen-cmd'


# 添加别名（选择合适的shell配置文件，这里以bash_profile举例）
echo "alias ac='aicli chat'" >> ~/.bash_profile
# 重新加载shell配置
source ~/.bash_profile
# 验证别名
ac
```

## 贡献指南
如果您在使用或开发过程中遇到问题，欢迎在Issues页面提交问题或讨论。欢迎任何形式的贡献！

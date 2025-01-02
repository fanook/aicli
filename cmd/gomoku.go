package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	openapi "github.com/fanook/aicli/internal/openapi"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	boardSize = 15

	// ANSI color codes
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m" // AI
	ColorGreen = "\033[32m" // Human
)

// Player 类型
type Player int

const (
	Human Player = iota
	AIPlayer
)

// Game 结构体
type Game struct {
	board     [][]rune
	current   Player
	moveCount int
}

// 创建新游戏
func newGame() *Game {
	board := make([][]rune, boardSize)
	for i := range board {
		board[i] = make([]rune, boardSize)
		for j := range board[i] {
			board[i][j] = '·' // 空位用 · 表示
		}
	}
	return &Game{
		board:   board,
		current: Human,
	}
}

var gomokuCmd = &cobra.Command{
	Use:     "gomoku",
	Short:   "与 AI 一起玩五子棋",
	Long:    `在终端中与 AI 对战，进行五子棋游戏。`,
	Example: `  aicli gomoku`,
	Run: func(cmd *cobra.Command, args []string) {
		game := newGame()
		game.play()
	},
}

func init() {
	rootCmd.AddCommand(gomokuCmd)
}

func (g *Game) play() {
	reader := bufio.NewReader(os.Stdin)
	for {
		g.renderBoard()
		if g.current == Human {
			fmt.Print("你 (X) 的回合。输入格式如 'H8'，或输入 'exit' 退出游戏: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				logrus.Fatalf("读取输入失败: %v", err)
			}
			input = strings.TrimSpace(input)
			if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" {
				fmt.Println("游戏结束。再见！")
				break
			}
			x, y, err := parseMove(input)
			if err != nil {
				fmt.Println("无效的输入，请重新输入。")
				continue
			}
			if !g.placeStone(x, y, 'X') {
				fmt.Println("该位置已被占用，请重新输入。")
				continue
			}
			logrus.Infof("玩家移动: %s", input)
			if g.checkWin(x, y, 'X') {
				g.renderBoard()
				fmt.Println("🎉 你赢了！🎉")
				break
			}
			g.current = AIPlayer
		} else {
			fmt.Println("AI (O) 正在思考...")
			x, y, reason, err := g.aiMoveWithRetry(3)
			if err != nil {
				fmt.Println("AI 发生错误:", err)
				break
			}
			fmt.Printf("AI 下在 %s\n", moveToString(x, y))
			fmt.Printf("AI 的决策理由: %s\n", reason)
			logrus.Infof("AI移动: %s, 理由: %s", moveToString(x, y), reason)
			if g.checkWin(x, y, 'O') {
				g.renderBoard()
				fmt.Println("😞 AI 赢了！")
				break
			}
			g.current = Human
		}
		g.moveCount++
		if g.moveCount >= boardSize*boardSize {
			g.renderBoard()
			fmt.Println("平局！")
			break
		}
	}
}

// 渲染棋盘
func (g *Game) renderBoard() {
	// 打印顶部列号
	fmt.Print("\n  ")
	for i := 0; i < boardSize; i++ {
		fmt.Printf(" %2d", i+1)
	}
	fmt.Println()

	// 打印每一行
	for i := 0; i < boardSize; i++ {
		fmt.Printf("%2s ", string(rune('A'+i)))
		for j := 0; j < boardSize; j++ {
			switch g.board[i][j] {
			case 'X':
				fmt.Printf(" %sX%s ", ColorGreen, ColorReset) // 使用绿色显示玩家
			case 'O':
				fmt.Printf(" %sO%s ", ColorRed, ColorReset) // 使用红色显示AI
			default:
				fmt.Printf(" %c ", g.board[i][j])
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

// 解析移动
func parseMove(move string) (int, int, error) {
	move = strings.TrimSpace(strings.ToUpper(move))
	if len(move) < 2 || len(move) > 3 {
		return 0, 0, fmt.Errorf("invalid move format")
	}

	rowChar := string(move[0])
	if rowChar < "A" || rowChar > string(rune('A'+boardSize-1)) {
		return 0, 0, fmt.Errorf("invalid row")
	}
	row := int(rune(rowChar[0]) - 'A')

	var column int
	_, err := fmt.Sscanf(move[1:], "%d", &column)
	if err != nil || column < 1 || column > boardSize {
		return 0, 0, fmt.Errorf("invalid column")
	}

	return row, column - 1, nil
}

// 将移动转换为字符串
func moveToString(x, y int) string {
	return fmt.Sprintf("%c%d", rune('A'+x), y+1)
}

// 放置棋子
func (g *Game) placeStone(x, y int, stone rune) bool {
	if x < 0 || x >= boardSize || y < 0 || y >= boardSize {
		return false
	}
	if g.board[x][y] != '·' {
		return false
	}
	g.board[x][y] = stone
	return true
}

// 检查是否获胜
func (g *Game) checkWin(x, y int, stone rune) bool {
	directions := [][]int{
		{1, 0},  // 水平
		{0, 1},  // 垂直
		{1, 1},  // 主对角线
		{1, -1}, // 副对角线
	}
	for _, dir := range directions {
		count := 1
		// 向正方向延伸
		for i := 1; i < 5; i++ {
			nx, ny := x+dir[0]*i, y+dir[1]*i
			if nx < 0 || nx >= boardSize || ny < 0 || ny >= boardSize || g.board[nx][ny] != stone {
				break
			}
			count++
		}
		// 向反方向延伸
		for i := 1; i < 5; i++ {
			nx, ny := x-dir[0]*i, y-dir[1]*i
			if nx < 0 || nx >= boardSize || ny < 0 || ny >= boardSize || g.board[nx][ny] != stone {
				break
			}
			count++
		}
		if count >= 5 {
			return true
		}
	}
	return false
}

func (g *Game) boardToString() (string, error) {
	type Position struct {
		Row    string `json:"row"`
		Column int    `json:"column"`
		Stone  string `json:"stone"`
	}

	var positions []Position
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			stone := string(g.board[i][j])
			if stone != "·" {
				positions = append(positions, Position{
					Row:    string('A' + i),
					Column: j + 1,
					Stone:  stone,
				})
			}
		}
	}

	boardJSON, err := json.MarshalIndent(positions, "", "  ")
	if err != nil {
		return "", fmt.Errorf("无法序列化棋盘状态为JSON: %v", err)
	}
	return string(boardJSON), nil
}

// AI走棋逻辑，通过OpenAI API获取移动和理由
func (g *Game) aiMove() (int, int, string, error) {
	apiKey := os.Getenv("AICLI_OPENAI_API_KEY")
	if apiKey == "" {
		return 0, 0, "", fmt.Errorf("未设置AICLI_OPENAI_API_KEY环境变量")
	}

	// 构建棋盘状态字符串
	boardStr, _ := g.boardToString()

	// 构建提示语，明确AI的角色和棋盘的意义
	prompt := fmt.Sprintf(
		`你是一个五子棋AI，使用棋子 'O'。你将在一个15x15的棋盘上与玩家（使用 'X'）对战。棋盘的行用字母A到O表示，列用数字1到15表示。以下是当前的棋盘状态，其中 'X' 代表玩家，'O' 代表AI，'·' 代表空位。

当前棋盘状态：
%s

请根据当前棋盘状态，我给你的棋盘状态中只有当前已落子的状态，其他都是空位。你需要重新构建这个2维棋盘，并运用以下策略选择一个最佳的移动位置：

1. **阻止玩家**：如果玩家有形成五连珠的潜在威胁，你应优先阻止。你需要观察每一个行、列、对角线上的棋子，防止玩家形成五连珠。你如果想阻止对角线上形成连接，你的棋子也应该落在对角线上。
2. **进攻AI**：如果你有形成五连珠的机会，应优先抓住。
3. **战略布局**：在棋盘中心或关键位置布局，增加未来的胜利机会。
4. **预测对手动作**：预测玩家的下一步可能的动作，提前做好防御或进攻准备。
5. **利用套路**：运用五子棋的常见套路和技巧，提升获胜的概率。

你只能选择空位。请严格按照以下格式返回你的移动和理由，不要添加任何额外的说明：

reason: 这一步为什么这样走
path: H8

**请仅返回上述格式的内容，不要添加任何额外的说明。**`,
		boardStr,
	)

	// 调用 openapi.GenerateContent
	response, err := openapi.GenerateContent(apiKey, prompt)
	if err != nil {
		logrus.Errorf("调用OpenAI API失败: %v", err)
		return 0, 0, "", fmt.Errorf("调用OpenAI API失败: %v", err)
	}
	logrus.Infof("AI响应:\n%s", response)

	// 解析AI的响应
	reason, path, err := parseAIResponse(response)
	if err != nil {
		logrus.Errorf("解析AI响应失败: %v", err)
		return 0, 0, "", fmt.Errorf("解析AI响应失败: %v", err)
	}

	// 解析移动位置
	moveX, moveY, err := parseMove(path)
	if err != nil {
		logrus.Errorf("无法解析AI的移动位置: %v", err)
		return 0, 0, "", fmt.Errorf("无法解析AI的移动位置: %v", err)
	}

	// 检查AI移动的有效性
	if !g.placeStone(moveX, moveY, 'O') {
		logrus.Errorf("AI选择的位置已被占用: %s", path)
		return 0, 0, "", fmt.Errorf("AI选择的位置已被占用: %s", path)
	}

	logrus.Infof("AI选择的位置: %s, 理由: %s", path, reason)

	return moveX, moveY, reason, nil
}

// AI走棋逻辑，通过OpenAI API获取移动和理由，带重试机制
func (g *Game) aiMoveWithRetry(maxRetries int) (int, int, string, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		x, y, reason, err := g.aiMove()
		if err != nil {
			logrus.Errorf("AI第 %d 次尝试失败: %v", attempt, err)
			continue
		}
		return x, y, reason, nil
	}
	return 0, 0, "", fmt.Errorf("AI在 %d 次尝试后仍未选择有效的位置", maxRetries)
}

// parseAIResponse 解析AI响应，提取reason和path
func parseAIResponse(response string) (string, string, error) {
	// 优化后的正则表达式，优先匹配两位数列号
	reasonRegex := regexp.MustCompile(`(?i)reason:\s*(.*)`)
	pathRegex := regexp.MustCompile(`(?i)path:\s*([A-O]1[0-5]|[A-O][1-9])\b`)

	reasonMatch := reasonRegex.FindStringSubmatch(response)
	pathMatch := pathRegex.FindStringSubmatch(response)

	if len(reasonMatch) < 2 || len(pathMatch) < 2 {
		return "", "", fmt.Errorf("AI的响应格式不正确: %s", response)
	}

	reason := strings.TrimSpace(reasonMatch[1])
	path := strings.TrimSpace(pathMatch[1])

	return reason, path, nil
}

package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	openai "github.com/fanook/aicli/internal/openapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var processDataCmd = &cobra.Command{
	Use:   "process-data",
	Short: "从数据源中读取数据，调用AI生成回复，并更新每行的 prompt 和 result 字段",
	Long: `该命令从指定的数据源（数据库或CSV文件）中读取数据行，
要求每行必须包含 id、content、prompt、result 四个字段，其中 result 为空表示未处理。
根据参数传递的prompt或行prompt,填充content后，生成最终 prompt。
再调用 AI 接口生成回复，最后将生成结果写入result字段。`,
	Run: func(cmd *cobra.Command, args []string) {
		runProcessData()
	},
}

var (
	// 数据源类型：db 或 csv
	sourceType string
	// CSV 模式下的输入文件与输出文件路径
	csvFile string
	csvOut  string

	// DB 模式下的数据库连接参数（以 MySQL 为例）
	dbHost  string
	dbPort  string
	dbUser  string
	dbPass  string
	dbName  string
	dbTable string

	// 全局 AI处理提示语（模板），如果为空则每行使用自身的 prompt 字段
	promptText string
)

func init() {
	rootCmd.AddCommand(processDataCmd)
	processDataCmd.Flags().StringVarP(&sourceType, "source", "s", "csv", "数据源类型：db 或 csv, 默认为csv")
	processDataCmd.Flags().StringVarP(&csvFile, "csv-file", "f", "", "CSV输入文件路径（当数据源为csv时必填）")
	processDataCmd.Flags().StringVarP(&csvOut, "csv-out", "o", "output.csv", "CSV输出文件路径（当数据源为csv时有效）")

	// 数据库连接参数（当数据源为db时必填）
	processDataCmd.Flags().StringVar(&dbHost, "db-host", "localhost", "数据库主机地址")
	processDataCmd.Flags().StringVar(&dbPort, "db-port", "3306", "数据库端口")
	processDataCmd.Flags().StringVarP(&dbUser, "db-user", "u", "root", "数据库用户名")
	processDataCmd.Flags().StringVarP(&dbPass, "db-pass", "P", "", "数据库密码")
	processDataCmd.Flags().StringVar(&dbName, "db-name", "", "数据库名称")
	processDataCmd.Flags().StringVar(&dbTable, "db-table", "", "数据库表名")

	// 全局提示参数
	processDataCmd.Flags().StringVarP(&promptText, "prompt", "p", "", "全局AI处理提示语模板，若为空则使用每行数据中的 prompt 字段")
}

func runProcessData() {
	apiKey := os.Getenv("AICLI_OPENAI_API_KEY")
	if apiKey == "" {
		logrus.Fatal("未设置 AICLI_OPENAI_API_KEY 环境变量")
	}

	switch strings.ToLower(sourceType) {
	case "csv":
		if csvFile == "" {
			logrus.Fatal("当数据源为csv时，必须指定--csv-file参数")
		}
		processCSV(apiKey, promptText, csvFile, csvOut)
	case "db":
		if dbName == "" || dbTable == "" {
			logrus.Fatal("当数据源为db时，必须指定--db-name 和 --db-table 参数")
		}
		processDB(apiKey, promptText, dbHost, dbPort, dbUser, dbPass, dbName, dbTable)
	default:
		logrus.Fatalf("未知的数据源类型：%s", sourceType)
	}
}

// processCSV 从CSV文件中读取数据，调用AI生成回复后输出到新的CSV文件
// CSV 文件中每行必须有四列：id, content, prompt, result
func processCSV(apiKey, globalPrompt, inputFile, outputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		logrus.Fatalf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logrus.Fatalf("读取CSV文件失败: %v", err)
	}

	if len(records) == 0 {
		logrus.Info("CSV文件为空")
		return
	}

	header := records[0]
	if len(header) < 4 {
		logrus.Fatal("CSV表头列数不足，要求至少有 id, content, prompt, result 四列")
	}
	outputRecords := [][]string{header}

	var tmpl *template.Template
	if globalPrompt != "" {
		tmpl, err = template.New("prompt").Parse(globalPrompt)
		if err != nil {
			logrus.Fatalf("解析全局提示模板失败: %v", err)
		}
	}

	total := len(records) - 1

	for i, row := range records[1:] {
		if len(row) < 4 {
			logrus.Errorf("行 %d 列数不足，跳过", i+2)
			outputRecords = append(outputRecords, row)
			continue
		}

		id := row[0]
		content := row[1]
		existingPrompt := row[2]

		logrus.Infof("正在处理 [%d/%d] 行，ID: %s", i+1, total, id)

		var finalPrompt string
		var buf bytes.Buffer
		if tmpl == nil {
			tmpl, err = template.New("prompt").Parse(existingPrompt)
			if err != nil {
				logrus.Fatalf("解析行[id:%d]提示模板失败: %v", err, id)
			}
		}
		err = tmpl.Execute(&buf, struct {
			Content string
		}{
			Content: content,
		})
		if err != nil {
			logrus.Errorf("行 %d (ID:%s) 模板执行失败: %v", i+2, id, err)
			finalPrompt = existingPrompt
		} else {
			finalPrompt = buf.String()
		}

		logrus.Info("finalPrompt: ", finalPrompt)

		rowCtx, rowCancel := context.WithTimeout(context.Background(), 60*time.Second)
		replyChan := make(chan string, 1)
		errChan := make(chan error, 1)

		go func(prompt string) {
			reply, err := openai.GenerateContent(apiKey, prompt)
			if err != nil {
				errChan <- err
			} else {
				replyChan <- reply
			}
		}(finalPrompt)

		var reply string
		select {
		case <-rowCtx.Done():
			logrus.Errorf("行 %d (ID:%s) 处理超时", i+2, id)
			reply = ""
		case err := <-errChan:
			logrus.Errorf("行 %d (ID:%s) 生成回复失败: %v", i+2, id, err)
			reply = ""
		case reply = <-replyChan:
			logrus.Infof("行 %d (ID:%s) 回复生成成功", i+2, id)
		}
		rowCancel()

		row[3] = reply
		outputRecords = append(outputRecords, row)
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		logrus.Fatalf("创建输出CSV文件失败: %v", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	err = writer.WriteAll(outputRecords)
	if err != nil {
		logrus.Fatalf("写入CSV文件失败: %v", err)
	}
	writer.Flush()
	logrus.Infof("处理完成，输出文件: %s", outputFile)
}

// processDB 从数据库中读取未处理的数据行，调用AI生成回复后更新记录
// 数据库表必须有 id, content, prompt, result 四个字段，其中 result 为空表示未处理
func processDB(apiKey, globalPrompt, host, port, user, pass, dbName, tableName string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	var tmpl *template.Template
	if globalPrompt != "" {
		tmpl, err = template.New("prompt").Parse(globalPrompt)
		if err != nil {
			logrus.Fatalf("解析全局提示模板失败: %v", err)
		}
	}

	updateStmt, err := db.PrepareContext(context.Background(),
		fmt.Sprintf("UPDATE %s SET prompt = ?, result = ? WHERE id = ?", tableName))
	if err != nil {
		logrus.Fatalf("准备更新语句失败: %v", err)
	}
	defer updateStmt.Close()

	// 分页查询，每页处理 100 行
	pageSize := 100
	totalProcessed := 0
	page := 0
	startID := 0

	for {
		query := fmt.Sprintf("SELECT id, content, prompt, result FROM %s WHERE (result IS NULL OR result = '') AND id > %d ORDER BY id ASC LIMIT %d", tableName, startID, pageSize)
		rows, err := db.QueryContext(context.Background(), query)
		if err != nil {
			logrus.Fatalf("分页查询失败: %v", err)
		}

		recordsInPage := 0
		for rows.Next() {
			var id int
			var content, rowPrompt, result string
			if err := rows.Scan(&id, &content, &rowPrompt, &result); err != nil {
				logrus.Errorf("扫描数据失败: %v", err)
				continue
			}
			recordsInPage++
			totalProcessed++
			logrus.Infof("正在处理记录 [%d] (总计 %d) - ID: %d", recordsInPage+(page*pageSize), totalProcessed, id)

			startID = id

			var finalPrompt string
			if globalPrompt != "" {
				var buf bytes.Buffer
				err = tmpl.Execute(&buf, struct {
					Content string
				}{
					Content: content,
				})
				if err != nil {
					logrus.Errorf("ID %d 模板执行失败: %v", id, err)
					finalPrompt = rowPrompt
				} else {
					finalPrompt = buf.String()
				}
			} else {
				finalPrompt = rowPrompt
			}

			rowCtx, rowCancel := context.WithTimeout(context.Background(), 60*time.Second)
			replyChan := make(chan string, 1)
			errChan := make(chan error, 1)
			go func(prompt string) {
				reply, err := openai.GenerateContent(apiKey, prompt)
				if err != nil {
					errChan <- err
				} else {
					replyChan <- reply
				}
			}(finalPrompt)

			var reply string
			select {
			case <-rowCtx.Done():
				logrus.Errorf("ID %d 处理超时", id)
				reply = ""
			case err := <-errChan:
				logrus.Errorf("ID %d 生成回复失败: %v", id, err)
				reply = ""
			case reply = <-replyChan:
				_, err = updateStmt.ExecContext(context.Background(), finalPrompt, reply, id)
				if err != nil {
					logrus.Errorf("ID %d 更新结果失败: %v", id, err)
				} else {
					logrus.Infof("ID %d 更新成功", id)
				}
			}
			rowCancel()
		}
		rows.Close()

		if recordsInPage == 0 {
			logrus.Infof("没有更多记录需要处理，总共更新 %d 条记录", totalProcessed)
			break
		}

		page++
		logrus.Infof("第 %d 页处理完成，共处理 %d 条记录", page, totalProcessed)
	}
}

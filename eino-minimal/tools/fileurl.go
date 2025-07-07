package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/unidoc/unioffice/document"
)

type FileUrlParams struct {
	File_url string `json:"file_url" binding:"required"`
}

// 注册模型工具  暂时没用
func FileUrlTool() tool.InvokableTool {
	// info := &schema.ToolInfo{
	// 	Name: "Handle FileURL",
	// 	Desc: "从URL中获取文件内容并解析",
	// 	ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
	// 		"file_url": {
	// 			Desc:     "文件URL",
	// 			Type:     schema.String,
	// 			Required: true,
	// 		},
	// 	}),
	// }
	// return utils.NewTool(info, ParseFileURL)
	fileurlTool := utils.NewTool(&schema.ToolInfo{
		Name: "ParseFileURL",
		Desc: "从URL中获取文件内容并解析",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"file_url": &schema.ParameterInfo{
					// Desc:     "文件URL",
					Type:     schema.String,
					Required: true,
				},
			}),
	}, ParseFileURL)
	return fileurlTool
}

func ParseFileURL(_ context.Context, params *FileUrlParams) (string, error) {
	file_url := params.File_url
	fmt.Println(file_url, "\t FIleURLParser工具被调用了")
	// url := fmt.Sprintf("https://cn.apihz.cn/api/tianqi/tqyb.php?id=88888888&key=88888888&sheng=%s&place=%s", sheng, place)
	// 1. 从url下载文件
	resp, err := http.Get(file_url)
	if err != nil {
		return "", fmt.Errorf("下载文件失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("resp.Body关闭失败: %v", err)
		}
		log.Printf("ParseFileURL工具调用完成")
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败: %s", resp.Status)
	}
	log.Printf("从url下载文件成功完成")

	// 目前测试txt不需要判断文件类型 // 文件类型
	// ext := filepath.Ext(file_url)
	// content, err := parseFileContent(ext, resp.Body)
	// if err != nil {
	// 	fmt.Printf("parse Content Failed: %v", err)
	// 	return "", nil
	// }
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取文件内容失败:%v", err)
	}
	return string(content), nil
}

func TestParseFileURL(file_url string) (string, error) {
	fp := &FileUrlParams{}
	ctx := context.Background()
	// fp.File_url = "https://filesamples.com/samples/document/doc/sample2.doc"
	fp.File_url = file_url
	ctt, err := ParseFileURL(ctx, fp)
	if err != nil {
		return "", err
	}
	return ctt, err
}

// resp.Body io.ReadCLoser 实现了io.Reader的接口，直接传入
func TestpdfParser(ctx context.Context, reader io.Reader) {
	//  with URL of pdf file
	pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{})
	if err != nil {
		log.Printf("pdf解析器 创建失败: %v", err)
		return
	}
	docs, err := pdfParser.Parse(ctx, reader)
	if err != nil {
		log.Printf("pdf解析器 解析失败")
		return
	}
	for _, doc := range docs {
		fmt.Println(doc.Content)
	}
}

func TestLocalPDFParser(filepath string) error {
	ctx := context.Background()
	p, err := pdf.NewPDFParser(ctx, &pdf.Config{
		ToPages: false, // 不按页面分割
	})
	if err != nil {
		log.Panicf("%v", err)
	}
	// 打开pdf
	file, err := os.Open("document.pdf")
	if err != nil {
		log.Panicf("%v", err)
	}
	defer file.Close()

	// 解析文档
	docs, err := p.Parse(ctx, file,
		parser.WithURI("document.pdf"),
		parser.WithExtraMeta(map[string]any{
			"source": "./document.pdf",
		}),
	)
	if err != nil {
		log.Panicf("%v", err)
	}
	for _, doc := range docs {
		fmt.Println(doc.Content)
	}
	return nil
}

// parseFileContent 根据文件类型解析内容  WIP for other type
func parseFileContent(filePath string, body io.ReadCloser) (string, error) {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".doc", ".docx":
		return parseDocFile(body)
	case ".txt":
		return parserTxtFile(filePath, body)
	default:
		return "", fmt.Errorf("不支持的文件类型: %s", ext)
	}
}

func parserTxtFile(file_url string, body io.ReadCloser) (string, error) {
	// 1. 从url下载文件
	resp, err := http.Get(file_url)
	if err != nil {
		return "", fmt.Errorf("下载文件失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("resp.Body关闭失败: %v", err)
		}
		log.Printf("ParseTxtFile调用完成")
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败: %s", resp.Status)
	}
	log.Printf("从url下载文件成功完成")

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取文件内容失败:%v", err)
	}
	return string(content), nil
}

func parseDocFile(body io.ReadCloser) (string, error) {
	// 1. 读取响应内容到字节流
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		return "", fmt.Errorf("读取内容失败: %v", err)
	}

	// 2. 解析DOC/DOCX文件
	doc, err := document.Read(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return "", fmt.Errorf("解析文档失败: %v", err)
	}

	// 3. 提取纯文本
	var content string
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			content += run.Text()
		}
		content += "\n" // 段落换行
	}
	return content, nil
}

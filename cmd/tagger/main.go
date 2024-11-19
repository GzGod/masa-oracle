/*
主包 - 打标签工具

打标签工具是一个实用工具，旨在自动化版本控制、构建和为 Go 项目打标签的过程。它通过将多个步骤封装成一个命令来简化工作流程，减少人为错误的可能性并提高效率。

打标签工具执行以下操作：

1. 更新项目：它导航到项目目录并从指定分支拉取最新更改。然后使用 'go mod tidy' 整理依赖项。

2. 推送更改：它暂存并提交带有用户提供的提交消息的更改，然后将提交推送到指定分支。

3. 为项目打标签：它使用用户提供的版本标签为当前提交打标签，并将标签推送到远程仓库。

4. 构建项目：它为多个平台（Linux、Darwin 和 Windows）构建项目，并将二进制文件输出到项目目录中的 'bin' 目录。

打标签工具从命令行调用，带有四个标志：'projectDir'（项目目录）、'version'（要打的版本标签）、'branch'（要使用的分支）和 'commitMessage'（提交消息）。必须提供所有标志。

使用示例：
go run tag_project.go -projectDir=/path/to/project -version=1.0.0 -branch=main -commitMessage="Update go.mod"

该实用工具适用于需要定期版本控制、构建和为项目打标签的开发人员。它假定用户有一个可用的 Go 环境并熟悉基本的 Git 操作。
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func 更新项目(项目目录, 分支 string) error {
	// 切换到项目目录
	if err := os.Chdir(项目目录); err != nil {
		return err
	}

	// 执行 git pull
	if err := 运行命令("git", "pull", "origin", 分支); err != nil {
		return err
	}
	// 执行 go mod tidy
	if err := 运行命令("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("[-] 整理依赖项失败: %w", err)
	}
	return nil
}

func 推送更改(项目目录, 提交消息, 分支 string) (err error) {
	// 切换到项目目录
	if err := os.Chdir(项目目录); err != nil {
		return err
	}

	// 执行 git add .
	if err := 运行命令("git", "add", "."); err != nil {
		return err
	}

	// 执行 git commit -m "更新 go.mod"
	if err := 运行命令("git", "commit", "-m", 提交消息); err != nil {
		return err
	}

	// 执行 git push origin main
	if err := 运行命令("git", "push", "origin", 分支); err != nil {
		return err
	}
	return nil
}

func 为项目打标签(项目目录, 版本 string) error {
	// 切换到项目目录
	if err := os.Chdir(项目目录); err != nil {
		return err
	}
	if err := 运行命令("git", "tag", "-a", 版本, "-m", fmt.Sprintf("版本 %s", 版本)); err != nil {
		return err
	}

	// 执行 git push --tags
	if err := 运行命令("git", "push", "--tags"); err != nil {
		return err
	}
	return nil
}

func 运行命令(命令 string, 参数 ...string) error {
	fmt.Printf("运行命令: %s %v\n", 命令, 参数)
	cmd := exec.Command(命令, 参数...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func 打标签项目(项目目录, 版本, 分支, 提交消息 string) (err error) {
	if 项目目录 == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal("无法找到用户的 home 目录")
		}
		项目目录 = filepath.Join(usr.HomeDir, "github", "masa-finance")
	}
	err = 更新项目(项目目录, 分支)
	if err != nil {
		return err
	}
	err = 推送更改(项目目录, 提交消息, 分支)
	if err != nil {
		return err
	}
	err = 为项目打标签(项目目录, 版本)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// 定义标志
	项目目录 := flag.String("projectDir", "", "项目目录")
	版本 := flag.String("version", "", "要打的版本标签")
	分支 := flag.String("branch", "", "要使用的分支")
	提交消息 := flag.String("commitMessage", "", "提交消息")

	// 解析标志
	flag.Parse()

	// 检查是否提供了所有标志
	if *项目目录 == "" || *版本 == "" || *分支 == "" || *提交消息 == "" {
		fmt.Println("错误: 必须提供所有标志")
		fmt.Println("用法: tag_project -projectDir=<projectDir> -version=<version> -branch=<branch> -commitMessage=<commitMessage>")
		os.Exit(1)
	}

	// 使用解析的参数调用打标签项目
	err := 打标签项目(*项目目录, *版本, *分支, *提交消息)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Println("更新和打标签成功完成。")
	}
}

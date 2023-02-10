/*
Copyright © 2023 enjoy <i@mcenjoy.cn>
*/
package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"io"
	"os"
	"os/exec"
	"qqwry/dat"
	"regexp"
	"time"
)

var (
	listPage        = "https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&album_id=2329805780276838401&f=json&count=10"
	innoExtractPath = "innoextract"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新数据库",
	Long:  `更新数据库`,
	RunE: func(cbCmd *cobra.Command, args []string) error {
		fmt.Printf("清理 缓存文件\n")
		err := os.RemoveAll("tmp")
		if err != nil {
			return err
		}
		err = os.Mkdir("tmp", 0777)
		if err != nil {
			return err
		}
		defer os.RemoveAll("tmp")
		qqwryPath, err := cbCmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		fs, err := os.Open(qqwryPath)
		downloadNew := false
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("未找到数据库文件，直接下载最新版本")
				downloadNew = true
			} else {
				return err
			}

		}
		var nowVersion *time.Time
		if !downloadNew {
			d := dat.New(fs)
			nowVersion, err = d.Version()
			fs.Close()
			if err != nil {
				return err
			}
		}

		resp, err := req.R().Get(listPage)
		if err != nil {
			return err
		}
		list := gjson.ParseBytes(resp.Bytes())
		if ret := list.Get("base_resp.ret"); !ret.Exists() || ret.Int() != 0 {
			panic("接口出现问题，无法获取！")
		}

		latest := list.Get("getalbum_resp.article_list.0")
		title := latest.Get("title").String()
		title = title[27 : len(title)-1]

		latestTime, err := time.ParseInLocation("2006-01-02", title, loc)
		if err != nil {
			return err
		}
		u := latest.Get("url").String()
		if !downloadNew {
			if latestTime.Sub(*nowVersion).Hours() > 0 {
				fmt.Printf("当前版本发布时间为:%s\n最新版本发布时间为:%s\n需要更新!\n", nowVersion.Format("20060102"), latestTime.Format("20060102"))
			} else {
				fmt.Printf("当前版本发布时间为:%s\n最新版本发布时间为:%s\n无需更新\n", nowVersion.Format("20060102"), latestTime.Format("20060102"))
				return nil
			}
		}

		fmt.Println("开始更新")
		resp, err = req.R().Get(u)
		if err != nil {
			return err
		}
		r := regexp.MustCompile(`https://www.cz88.net/soft/.*?\.zip`)
		result := r.FindString(resp.String())
		if result == "" {
			return fmt.Errorf("not found url")
		}
		fmt.Printf("下载地址为:%s\n", result)
		resp, err = req.R().Get(result)
		if err != nil {
			return err
		}

		zipReader, err := zip.NewReader(bytes.NewReader(resp.Bytes()), int64(len(resp.Bytes())))
		if err != nil {
			return err
		}

		setup, err := zipReader.Open("setup.exe")
		if err != nil {
			return err
		}

		fs1, err := os.OpenFile("tmp/qqwry.tmp", os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			return err
		}
		_, err = io.Copy(fs1, setup)
		fs1.Close()
		if err != nil {
			return err
		}
		fmt.Printf("下载完成，保存为qqwry.tmp\n")
		fmt.Printf("开始解压qqwry.dat\n")
		cmd := exec.Command(innoExtractPath, "-d", "./tmp", "-I", "app\\qqwry.dat", "tmp/qqwry.tmp")
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		err = os.Rename("tmp/app/qqwry.dat", qqwryPath)
		if err != nil {
			return err
		}
		fmt.Printf("qqwry.dat更新完毕\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&innoExtractPath, "iepath", "i", "innoextract", "innoExtract Path")
}

var loc = time.FixedZone("CST", 3600*8)

/*
Copyright © 2023 enjoy <i@mcenjoy.cn>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"os"
	"qqwry/dat"
	"strconv"
	"strings"
	"time"
)

var (
	memory = false
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search ipv4",
	Aliases: []string{"sc"},
	Short:   "查询ip归属地",
	Long:    `查询ip归属地`,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		qqwryPath, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}

		fs, err := os.Open(qqwryPath)
		if err != nil {
			return err
		}
		d := dat.New(fs, memory, func() {
			if memory {
				fmt.Println("关闭数据库文件句柄")
				fs.Close()
			}
		})
		if !memory {
			defer fs.Close()
		}
		t := time.Now()

		for _, ip_s := range args {
			ip := strings.Split(ip_s, ".")
			if len(ip) != 4 {
				return fmt.Errorf("ip error")
			}
			var m [4]byte
			for i, v := range ip {
				tmp, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				if tmp < 0 || tmp > 255 {
					return fmt.Errorf("ip error")
				}
				m[i] = byte(tmp)
			}
			country, area, err := d.SearchIp(net.IPv4(m[0], m[1], m[2], m[3]))
			if err != nil {
				return err
			}
			fmt.Printf("%s - %s %s \n", ip_s, country, area)
		}
		cost := time.Since(t).Milliseconds()
		fmt.Printf("查询耗时 %d ms\n", cost)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVarP(&memory, "memory", "m", false, "将数据库加载入内存")
}

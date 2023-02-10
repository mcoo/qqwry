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
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search ipv4",
	Aliases: []string{"sc"},
	Short:   "查询ip归属地",
	Long:    `查询ip归属地`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		qqwryPath, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		ip_s := args[0]
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
		fs, err := os.Open(qqwryPath)
		if err != nil {
			return err
		}
		defer fs.Close()
		d := dat.New(fs)
		country, area, err := d.SearchIp(net.IPv4(m[0], m[1], m[2], m[3]))
		if err != nil {
			return err
		}
		fmt.Printf("%s - %s %s\n", ip_s, country, area)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

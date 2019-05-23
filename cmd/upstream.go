// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/asoorm/go-bench-suite/upstream"

	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

var listenAddr, tlsCertFile, tlsKeyFile string

var upstreamCmd = &cobra.Command{
	Use:   "upstream",
	Short: "Starts a mock upstream server",
	Long:  `Starts a mock upstream server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("upstream called")

		if tlsCertFile != "" && tlsKeyFile != "" {
			log.Fatal(upstream.ServeTLS(listenAddr, tlsCertFile, tlsKeyFile))
		} else {
			log.Fatal(upstream.Serve(listenAddr))
		}
	},
}

func init() {
	rootCmd.AddCommand(upstreamCmd)

	upstreamCmd.PersistentFlags().StringVar(&listenAddr, "addr", ":8000", "Listen address for the server")
	upstreamCmd.PersistentFlags().StringVar(&tlsCertFile, "tlsCert", "", "Location of TLS cert file")
	upstreamCmd.PersistentFlags().StringVar(&tlsKeyFile, "tlsKey", "", "Location of TLS key file")
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "새로운 삼평 하이-로우 (PLS7) 게임을 시작합니다.",
	Long:  `새로운 삼평 하이-로우 (PLS7) 게임을 시작합니다. 1명의 플레이어와 5명의 CPU가 참여합니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1단계 목표: 환영 메시지 출력
		fmt.Println("=======================================================")
		fmt.Println("    삼평 하이-로우 (Pot Limit Sampyong 7 or better)")
		fmt.Println("=======================================================")
		fmt.Println("\n게임을 시작합니다!")
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	// 여기에 로컬 플래그 등을 추가할 수 있습니다.
	// 예: playCmd.Flags().IntP("players", "p", 6, "참여할 플레이어 수")
}

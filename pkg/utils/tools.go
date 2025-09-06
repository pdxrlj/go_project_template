package utils

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

// 请求消耗的时间
func TimeElapsedCostStrColor(elapsed float64) string {
	if elapsed < 1 {
		return color.GreenString("%.2fms", elapsed)
	} else if elapsed < 10 {
		return color.YellowString("%.2fms", elapsed)
	} else {
		return color.RedString("%.2fs", elapsed)
	}
}

func PP(tag string, v any) {
	fmt.Print(tag, " ")
	json, err := json.MarshalIndent(v, "", "  	")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(json))
}

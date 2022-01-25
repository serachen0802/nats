package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	url := nats.DefaultURL
	if os.Getenv("IP") != "" {
		url = "nats://" + os.Getenv("IP") + ":4222"
	}
	con, err := nats.Connect(url)
	if err != nil {
		panic(err)
	}
	js, err := con.JetStream()
	if err != nil {
		panic(err)
	}

	// 依照所傳進來的參數決定pub sub

	//pub(沒有帶任何檔名 所以pub資料)
	if len(os.Args) > 1 {
		// pub 依照檔名做pub

		fileName := os.Args[1]
		fmt.Printf("fileName: %v\n", fileName)

		value, err := os.ReadFile(fileName)
		if err != nil {
			panic(err)
		}

		msg := nats.NewMsg("tt1")
		max := 30000
		// part := 0
		// fmt.Println(len(value))
		// total := len(value) / max
		// if len(value)%max > 0 {
		// 	total++
		// }
		buf := bytes.NewBuffer(value)

		ttt := [][]byte{}
		for {
			bb := buf.Next(max)
			if len(bb) == 0 {
				break
			}
			ttt = append(ttt, bb)
			// fmt.Println(part)
			// fmt.Println(total)
		}

		for part, bb := range ttt {
			// part++

			msg.Data = bb
			msg.Header.Set("part", strconv.Itoa(part))
			msg.Header.Set("total", strconv.Itoa(len(ttt)))
			// 預計接收後的儲存檔名
			msg.Header.Set("fileName", fileName)
			err = con.PublishMsg(msg)

			if err != nil {
				panic(err)
			}
		}

		// 拆圖片成幾部分

		// max := 2<<10 - 1024 // 1kb
		//43244
		// for i := 0; i < count; i++ {

		// }

		// n := len(value) / max
		// if len(value)%max > 0 {
		// 	n++
		// }
		// for i := 0; i < n; i++ {
		// 	value[max*i : max*(i+1)] // [0:1024], [1024:2048]
		// }

		// 檢查錯誤版本 - Zuolar
		// err = con.Publish("tt1", []byte(value))
		// if err != nil {
		// 	panic(err)
		// }
		//sub
	} else {
		sub1, err := js.SubscribeSync("tt1", nats.Bind("db1", "tt1"))
		if err != nil {
			panic(err)
		}
		defer sub1.Drain()
		defer sub1.Unsubscribe()

		// 取完全部圖片後 組合成完整圖片
		for {
			temp, err := sub1.NextMsg(3 * time.Second)
			if err != nil {
				panic(err)
			}

			part := temp.Header.Get("part")

			// 將拆解的圖片寫入temp底下
			err = os.WriteFile("temp/"+part, temp.Data, os.ModePerm)
			if err != nil {
				panic(err)
			}
			//temp下的檔案數量 = total 就可以儲存
			file, err := os.ReadDir("temp")
			if err != nil {
				panic(err)
			}

			// total:總共應該取得碎片數量
			total, err := strconv.Atoi(temp.Header.Get("total"))
			if err != nil {
				panic(err)
			}

			if total == len(file) {
				fmt.Println("可以合成圖片了")
				dir := "temp"
				Nbuf := bytes.NewBuffer(nil)

				for i := 0; i < total; i++ {
					searchFile := filepath.Join(dir, strconv.Itoa(i))
					file, err := os.ReadFile(searchFile)
					if err != nil {
						panic(err)
					}
					Nbuf.Write(file)
					//讀檔案 組合
				}
				fileName := temp.Header.Get("fileName")
				newDir := "new"
				NewPath := filepath.Join(newDir, fileName)
				err = os.WriteFile(NewPath, Nbuf.Bytes(), os.ModePerm)
				if err != nil {
					panic(err)
				}
				// 刪除temp下的檔案
				deleteFile, err := os.ReadDir("temp/")
				if err != nil {
					panic(err)
				}
				for _, d := range deleteFile {
					err := os.Remove("temp/" + d.Name())
					if err != nil {
						panic(err)
					}
				}
			}
		}

		// part := temp.Header.Get("part")
		// total := temp.Header.Get("total")

		// 判斷為完整圖片後做儲存
		// err = os.WriteFile(fileName, temp.Data, os.ModePerm)

		// fmt.Println(part)

		// }

		// sub1, err := js.Subscribe("tt1", func(msg *nats.Msg) {
		// 	if msg.Data != nil {
		// 		fmt.Println("save success")
		// 	}
		// }, nats.Bind("db1", "tt1"))
		// if err != nil {
		// 	panic(err)
		// }
		time.Sleep(3 * time.Second)
	}

	// sub := "123"

	// sub1, err := js.Subscribe(sub, func(msg *nats.Msg) {
	// 	fmt.Printf("msg: %v\n", msg.Data)
	// }, nats.Bind("str", "con"))

	con.Close()
}

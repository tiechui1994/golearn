package concurrent

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golearn/concurrent/ext"
)

func ErrGroup() {
	urls := []string{
		"http://www.exple.dtd",
		"qq", "zzz",
		"pp", "rwrw", "pwwjq", "wqwiwjkq",
		"http://www.baidu.vv",
		"http://www.qq.com"}

	var group *ext.Group
	group, _ = ext.WithContext(context.Background())
	for i := range urls {
		url := urls[i]
		group.Go(func() error {
			time.Sleep(1000 * time.Millisecond)
			response, err := http.Get(url)
			if err != nil {

				fmt.Printf("url:%v, err:%v\n", url, err)
				return err
			}

			fmt.Printf("url:%v code:%v \n", url, response.Status)
			return nil
		})
	}

	err := group.Wait()
	fmt.Printf("result:%v \n", err)
}

func WaitGroup() {
	urls := []string{
		"http://www.exple.dtd",
		"qq", "zzz",
		"pp", "rwrw", "pwwjq", "wqwiwjkq",
		"http://www.baidu.vv",
		"http://www.qq.com"}

	var group sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(len(urls))
	for i := range urls {
		url := urls[i]

		go func(ctx context.Context, url string) {
			defer group.Done()
			response, err := http.Get(url)
			if err != nil {
				cancel()
				fmt.Printf("url:%v, err:%v\n", url, err)
				return
			}

			if ctx.Err() != nil {
				return
			}

			fmt.Printf("url:%v code:%v \n", url, response.Status)
		}(ctx, url)
	}

	group.Wait()
}

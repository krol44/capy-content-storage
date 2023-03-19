package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestStress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			resExec, err := exec.Command("curl",
				"-H", "Token:some-token",
				"-H", "Storage:langlija",
				"-F", "file=@test-img.png",
				"--insecure", "http://localhost:8017").Output()

			if err != nil {
				t.Error(err)
			}

			var result Result
			err = json.Unmarshal(resExec, &result)
			if err != nil {
				t.Error(err)
			}

			if result.Status != true {
				t.Error("status upload not success")

			}
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	}(ctx)

	go func() {
		for {
			resExec, err := exec.Command("curl",
				"-H", "Token:some-token",
				"-H", "Storage:langlija",
				"-H", "Content-Type:application/json",
				"-d", `{"paths": []}`,
				"--insecure", "http://localhost:8017/files").Output()
			if err != nil {
				t.Error(err)
			}

			var result Files
			err = json.Unmarshal(resExec, &result)
			if err != nil {
				t.Error(err)
			}

			if result.Status != true {
				t.Error("get stats not success")

			}
			t.Logf("size: %dMB items: %d", result.Size>>20, result.Items)
			time.Sleep(time.Second)

		}
	}()

	time.Sleep(time.Second * 10)

	resExec, err := exec.Command("curl",
		"-H", "Token:some-token",
		"-H", "Storage:langlija",
		"-H", "Content-Type:application/json",
		"-d", `{"paths": ["all"]}`,
		"--insecure", "http://localhost:8017/files").Output()
	if err != nil {
		t.Error(err)
	}

	var result Files
	err = json.Unmarshal(resExec, &result)
	if err != nil {
		t.Error(err)
	}
	if result.Status != true {
		t.Error("get all not success")
	}

	for _, val := range result.Files {
		resExec, err := exec.Command("curl",
			"-H", "Token:some-token",
			"-H", "Storage:langlija",
			"-H", "Content-Type:application/json",
			"-d", fmt.Sprintf(`{"path": "%s"}`, val.Filename),
			"--insecure", "http://localhost:8017/remove").Output()
		if err != nil {
			t.Error(err)
		}

		var resultRem Files
		err = json.Unmarshal(resExec, &resultRem)
		if err != nil {
			t.Error(err)
		}
		if resultRem.Status != true {
			t.Error("remove not success")
		}
	}

	cancel()

	time.Sleep(time.Second)

	err = os.RemoveAll("files/")
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll("files-removed/")
	if err != nil {
		t.Error(err)
	}

	err = os.Mkdir("files", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	err = os.Mkdir("files-removed", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
}

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/grpc"
	"gopkg.in/bblfsh/sdk.v1/protocol"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Arguments: language sourceFile/s")
		os.Exit(1)
	}

	lang := os.Args[1]

	fmt.Println("Starting client...")
	cli, err := getClient("localhost:9432")
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")

	for _, f := range os.Args[2:] {
		ex, err := exists(f)
		if err != nil {
			panic(err)
		}

		if !ex {
			fmt.Println("File does not exist:", f)
			os.Exit(1)
		}

		fmt.Println("Parsing file:", f)
		err = parseFile(f, lang, cli)
		if err != nil {
			panic(err)
		}
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func getClient(endpoint string) (protocol.ProtocolServiceClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithTimeout(time.Second*2), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	cli := protocol.NewProtocolServiceClient(conn)
	return cli, nil
}

func parseFile(filename string, lang string, cli protocol.ProtocolServiceClient) error {

	content, err := ioutil.ReadFile(filename)
	if (err != nil) {
		return err;
	}

	if len(content) == 0 {
		fmt.Print("")
		os.Exit(0)
	}
	return parse(lang, string(content), cli)
}

func parse(lang string, content string, cli protocol.ProtocolServiceClient) error {
	req := &protocol.ParseRequest{
		Language: lang,
		Content:  content,
	}
	_, err := cli.Parse(context.Background(), req)
	if err != nil {
		return err
	}

	//fmt.Print(res.String())
	return nil
}

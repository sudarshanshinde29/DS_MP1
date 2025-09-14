package main

import (
	"MP1/properties"
	grep "MP1/protoBuilds"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

func main() {
	propsPath := flag.String("props", "cluster.properties", "Path to properties file")
	mode := flag.String("mode", "lines", "lines or count")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: grpccoordinator -props file -mode lines|count -- <grep options>")
		os.Exit(2)
	}

	p, err := properties.Load(*propsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	//fmt.Fprintf(os.Stderr, "loaded properties from %s: %v\n", *propsPath, p)
	n := p.Int("no.of.machines", 0)
	if n <= 0 {
		fmt.Fprintln(os.Stderr, "no.of.machines missing or zero")
		os.Exit(1)
	}

	targets := make([]string, 0, n)
	labels := make([]string, 0, n)

	for i := 0; i < n; i++ {
		ip := p[fmtKey("peer.machine.ip", i)]
		port := p[fmtKey("peer.machine.port", i)]
		name := p[fmtKey("peer.machine.name", i)]
		if ip == "" || port == "" {
			continue
		}
		targets = append(targets, ip+":"+port)
		labels = append(labels, name)
	}

	req := &grep.SearchRequest{GrepOptions: args, Mode: *mode}

	var wg sync.WaitGroup
	wg.Add(len(targets))
	var total int64

	// DEBUG
	//fmt.Fprintf(os.Stderr, "targets=%v labels=%v args=%v mode=%s\n", targets, labels, args, *mode)
	overallStart := time.Now()
	for i, target := range targets {
		hostLabel := labels[i]
		go func(target, label string) {
			fmt.Printf("In goroutine for %s (%s)\n", label, target) // DEBUG
			defer wg.Done()
			workerStart := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				fmt.Fprintf(os.Stderr, "[%s] dial: %v\n", label, err)
				return
			}
			defer conn.Close()
			cli := grep.NewGrepServiceClient(conn)
			stream, err := cli.Search(ctx, req)
			fmt.Printf("[%s] sent search request %+v\n", label, req) // DEBUG
			if err != nil {
				fmt.Fprintf(os.Stderr, "[%s] search: %v\n", label, err)
				return
			}
			for {
				resp, err := stream.Recv()
				if err != nil {
					if err.Error() != "EOF" {
						fmt.Fprintf(os.Stderr, "[%s] recv error: %v\n", label, err)
					}
					break
				}
				fmt.Printf("[%s] got response: %+v\n", label, resp) // DEBUG
				if *mode == "count" {
					fmt.Printf("[%s] count=%d\n", label, resp.Count)
					atomic.AddInt64(&total, resp.Count)
					continue
				}
				fp := resp.FilePath
				if fp == "" {
					fp = label
				}
				fmt.Printf("[%s] %s:%s\n", label, filepath.Base(fp), resp.Log)
			}
			duration := time.Since(workerStart).Milliseconds()
			fmt.Fprintf(os.Stderr, "[%s] WORKER_MS=%d\n", label, duration)
		}(target, hostLabel)
	}
	wg.Wait()
	fmt.Fprintf(os.Stderr, "OVERALL_MS=%d\n", time.Since(overallStart).Milliseconds())
	if *mode == "count" {
		fmt.Printf("TOTAL_COUNT=%d\n", total)
	}
}

func fmtKey(prefix string, i int) string {
	return fmt.Sprintf("%s%d", prefix, i)
}

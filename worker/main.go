package worker

import (
	grep "MP1/protoBuilds"
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type server struct {
	grep.UnimplementedGrepServiceServer
	logDir     string
	glob       string
	workerHost string
}

func (s *server) Search(req *grep.SearchRequest, stream grep.GrepService_SearchServer) error {
	files, _ := filepath.Glob(filepath.Join(s.logDir, s.glob))
	if len(files) == 0 {
		return nil
	}

	if req.Mode == "count" {
		args := append([]string{"-H", "-c"}, req.GrepOptions...)
		args = append(args, files...)
		cmd := exec.CommandContext(stream.Context().(context.Context), "grep", args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		defer cmd.Wait()
		sum := int64(0)
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			line := sc.Text()
			if i := strings.LastIndexByte(line, ':'); i >= 0 {
				if n, err := strconv.Atoi(strings.TrimSpace(line[i+1:])); err == nil {
					sum += int64(n)
				}
			}
		}
		return stream.SendMsg(&grep.SearchResponse{Host: s.workerHost, Count: sum})
	}

	args := append([]string{"--line-buffered", "-H"}, req.GrepOptions...)
	args = append(args, files...)
	cmd := exec.CommandContext(stream.Context().(context.Context), "grep", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	sc := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		fp := ""
		if i := strings.IndexByte(line, ':'); i >= 0 {
			fp = line[:i]
			line = line[i+1:]
		}
		if err := stream.SendMsg(&grep.SearchResponse{Host: s.workerHost, FilePath: fp, Log: line}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	address := flag.String("addr", ":6000", "Listening port")
	logDir := flag.String("logdir", ".", "directory with logs")
	glob := flag.String("glob", "machine.*.log", "glob for log files")
	workerHost := flag.String("label", "", "worker host")
	flag.Parse()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		fmt.Println("Failed to listen: ", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	grep.RegisterGrepServiceServer(s, &server{logDir: *logDir, glob: *glob, workerHost: *workerHost})
	fmt.Println("Worker is listening on", *address)
	if err := s.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serve", err)
	}
}

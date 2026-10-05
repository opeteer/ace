package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultIdleTimeout is the maximum duration to wait without any incoming data (2 minutes)
const DefaultIdleTimeout = 120 * time.Second

// RunWithIdleTimeout executes a command, streaming its stdout and stderr live to the terminal.
// It allows long-running commands to proceed indefinitely as long as data keeps arriving,
// but terminates the process if no data arrives on stdout or stderr for the idleTimeout duration.
func RunWithIdleTimeout(cmd *exec.Cmd, idleTimeout time.Duration) error {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var lastActivity int64
	atomic.StoreInt64(&lastActivity, time.Now().UnixNano())

	var timedOut int32
	done := make(chan struct{})

	// Heartbeat / idle watcher goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				last := atomic.LoadInt64(&lastActivity)
				if time.Since(time.Unix(0, last)) > idleTimeout {
					atomic.StoreInt32(&timedOut, 1)
					if cmd.Process != nil {
						_ = cmd.Process.Kill()
					}
					return
				}
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout live
	go func() {
		defer wg.Done()
		streamPipe(stdoutPipe, os.Stdout, &lastActivity)
	}()

	// Stream stderr live
	go func() {
		defer wg.Done()
		streamPipe(stderrPipe, os.Stderr, &lastActivity)
	}()

	wg.Wait()
	close(done)

	waitErr := cmd.Wait()

	if atomic.LoadInt32(&timedOut) == 1 {
		return fmt.Errorf("process stalled: no data received for %v (connection stalled)", idleTimeout)
	}

	return waitErr
}

func streamPipe(r io.Reader, w io.Writer, lastActivity *int64) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			atomic.StoreInt64(lastActivity, time.Now().UnixNano())
			_, _ = w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

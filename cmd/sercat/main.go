package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/maargenton/go-cli"
	"github.com/maargenton/go-cli/pkg/option"
	"go.bug.st/serial"
)

func main() {
	cli.Run(&cli.Command{
		Handler:     &sercatCmd{},
		Description: "Open a serial port and link TX / RX to stdin / stdout",
	})
}

type sercatCmd struct {
	Port     string `opts:"arg:1, name:port" desc:"name of the port to open"`
	Baudrate uint32 `opts:"arg:2, name:baudrate, default:115200" desc:"baudrate to use for communication"`
	Format   string `opts:"arg:3, name:format, default:8N1" desc:"communication format"`

	// Timestamp bool `opts:"-t,--timestamp" desc:"prefix every line with elapsed time"`
	// Verbose   bool `opts:"-v,--verbose"   desc:"display additional information on startup"`
}

func (opts *sercatCmd) Version() string {
	return "sercat v0.1.0"
}

func (opts *sercatCmd) Complete(opt *option.T, partial string) []string {
	if opt.Position == 1 {
		var ports = opts.availablePorts()
		if len(ports) != 0 {
			return ports
		}
	}
	if opt.Position == 2 {
		return opts.availableSpeeds()
	}
	if opt.Position == 3 {
		return opts.availableFormats()
	}

	return cli.DefaultCompletion(opt, partial)
}

func (opts *sercatCmd) Run() error {
	var mode, errMode = opts.getSerialMode()
	if errMode != nil {
		return errMode
	}
	var port, errOpen = serial.Open(opts.Port, mode)
	if errOpen != nil {
		return fmt.Errorf("failed to open port '%v': %w", opts.Port, errOpen)
	}

	return opts.ForwardPort(context.Background(), port)
}

func (opts *sercatCmd) ForwardPort(ctx context.Context, port serial.Port) error {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT)

	var wg sync.WaitGroup

	// Reader, port -> stdout
	wg.Add(1)
	go func() {
		io.Copy(os.Stdout, port)
		wg.Done()
	}()

	// Writer, stdin -> port, not in wait group, not blocking process exit
	go func() {
		io.Copy(port, os.Stdin)
	}()

	<-signals
	port.Close()

	wg.Wait()

	return nil
}

// ---------------------------------------------------------------------------

func (opts *sercatCmd) getSerialMode() (*serial.Mode, error) {
	var mode = &serial.Mode{
		BaudRate: int(opts.Baudrate),
	}

	if len(opts.Format) < 3 {
		return nil, fmt.Errorf("invalid format '%v'", opts.Format)
	}

	if v, err := strconv.ParseInt(opts.Format[0:1], 10, 0); err != nil {
		return nil, fmt.Errorf("invalid data bits '%v'", opts.Format[0:1])
	} else {
		mode.DataBits = int(v)
	}

	switch opts.Format[1:2] {
	case "N":
		mode.Parity = serial.NoParity
	case "O":
		mode.Parity = serial.OddParity
	case "E":
		mode.Parity = serial.EvenParity
	case "M":
		mode.Parity = serial.MarkParity
	case "S":
		mode.Parity = serial.SpaceParity

	default:
		return nil, fmt.Errorf("invalid parity bit '%v'", opts.Format[1:2])
	}

	switch opts.Format[2:] {
	case "1":
		mode.StopBits = serial.OneStopBit
	case "1.5":
		mode.StopBits = serial.OnePointFiveStopBits
	case "2":
		mode.StopBits = serial.TwoStopBits
	default:
		return nil, fmt.Errorf("invalid stop bits '%v'", opts.Format[2:])
	}

	return mode, nil
}

func (opts *sercatCmd) availablePorts() []string {
	var ports, _ = serial.GetPortsList()
	return ports
}

func (opts *sercatCmd) availableSpeeds() []string {
	return []string{
		"1200", "1800", "2400", "4800", "7200", "9600",
		"14400", "19200", "28800", "38400", "57600", "76800",
		"115200", "230400",
	}
}

func (opts *sercatCmd) availableFormats() []string {
	var formats = []string{}
	for _, db := range []int{5, 6, 7, 8} {
		for _, pb := range []string{"N", "O", "E", "M", "S"} {
			for _, sb := range []string{"1", "1.5", "2"} {
				formats = append(formats, fmt.Sprintf("%v%v%v", db, pb, sb))
			}
		}
	}
	return formats
}

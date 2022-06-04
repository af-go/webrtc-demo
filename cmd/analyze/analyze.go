package analyze

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pion/stun"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/spf13/cobra"
)

var (
	filename      string
	enableDebug   bool
	availableOnly bool
	log           logr.Logger
)

const ()

// AnalyzeCmd analyze command, utils to check webrtc status
var AnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "analyze webrtc environment",
}

var StunStatusCmd = &cobra.Command{
	Use:   "stun-service-status",
	Short: "check public stun service status",
	Run: func(cmd *cobra.Command, args []string) {

		log = zap.New(func(o *zap.Options) {
			o.Development = enableDebug
		})
		log.V(1).Info("checking stun status:")
		stuns, err := load(filename, log)
		if err != nil {
			log.Error(err, "failed to load stun list from file", "file", filename)
			return
		}
		result := analyzeStunStatus(stuns, availableOnly, log)

		for index := range result {
			realAddress := result[index].RealAddress
			if result[index].RealAddress == "" {
				realAddress = "Not Available"
			}
			fmt.Printf("%-50s %s\n", result[index].StunAddress, realAddress)
		}
	},
}

func init() {
	AnalyzeCmd.AddCommand(StunStatusCmd)
	StunStatusCmd.Flags().StringVarP(&filename, "file", "f", "test/testdata/stun.txt", "stun filename")
	StunStatusCmd.Flags().BoolVar(&enableDebug, "enable-debug", false, "enable debug log")
	StunStatusCmd.Flags().BoolVar(&availableOnly, "available-only", true, "only include avaiable stun servers")
}

func analyzeStunStatus(addresses []string, availableOnly bool, log logr.Logger) []StunStatus {
	statusChan := make(chan StunStatus, len(addresses))
	defer close(statusChan)
	result := []StunStatus{}
	var wg sync.WaitGroup
	for index := range addresses {
		wg.Add(1)
		go func(addr string, c chan StunStatus, log logr.Logger) {
			defer wg.Done()
			c <- StunStatus{StunAddress: addr, RealAddress: checkStatus(addr, c, log)}
		}(addresses[index], statusChan, log)

	}
	for i := 0; i < len(addresses); i++ {
		r := <-statusChan
		if r.RealAddress == "" && availableOnly {
			continue
		}
		result = append(result, r)
	}
	wg.Wait()
	return result
}

func load(filename string, log logr.Logger) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var data []string
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

type StunStatus struct {
	StunAddress string
	RealAddress string
}

func checkStatus(addr string, channel chan StunStatus, log logr.Logger) string {
	c, err := stun.Dial("udp4", addr)
	if err != nil {
		log.V(1).Info("failed to dial stun server", "error", err, "address", addr)
		return ""
	}
	var xorAddr stun.XORMappedAddress
	address := ""
	if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
		if res.Error != nil {
			log.V(1).Info("received stun event with error", "error", res.Error, "address", addr)
			return
		}
		if err1 := xorAddr.GetFrom(res.Message); err1 != nil {
			log.V(1).Info("failed to get address from stun event", "error", err1, "address", addr)
			return
		}
		address = xorAddr.String()
	}); err != nil {
		log.V(1).Info("failed to execute stun bind request", "error", err, "address", addr)
		return ""
	}
	if err := c.Close(); err != nil {
		log.Error(err, "failed to close udp session", "address", addr)
		return ""
	}
	return address
}

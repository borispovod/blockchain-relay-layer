// +build integ

package app

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	testUtils "github.com/dfinance/dnode/helpers/tests/utils"
)

type MockDVM struct {
	server        *grpc.Server
	failExecution bool
	failResponse  bool
	execDelay     time.Duration
}

func (s *MockDVM) SetExecutionFail() { s.failExecution = true }
func (s *MockDVM) SetExecutionOK()   { s.failExecution = false }
func (s *MockDVM) SetResponseFail()  { s.failResponse = true }
func (s *MockDVM) SetResponseOK()    { s.failResponse = false }
func (s *MockDVM) SetExecutionDelay(dur time.Duration) {
	s.execDelay = dur
}
func (s *MockDVM) Stop() {
	if s.server != nil {
		s.server.Stop()
	}
}

func (s *MockDVM) PublishModule(ctx context.Context, in *vm_grpc.VMPublishModule) (*vm_grpc.VMExecuteResponse, error) {
	if s.failExecution {
		return nil, grpcStatus.Errorf(codes.Internal, "failing gRPC execution")
	}

	resp := &vm_grpc.VMExecuteResponse{}
	if !s.failResponse {
		resp = &vm_grpc.VMExecuteResponse{
			WriteSet:     nil,
			Events:       nil,
			GasUsed:      1,
			Status:       vm_grpc.ContractStatus_Discard,
			StatusStruct: nil,
		}
	}

	return resp, nil
}

func StartMockDVMService(listener net.Listener) *MockDVM {
	s := &MockDVM{
		execDelay: 100 * time.Millisecond,
	}

	server := grpc.NewServer()
	vm_grpc.RegisterVMModulePublisherServer(server, s)

	go func() {
		server.Serve(listener)
	}()
	s.server = server

	return s
}

func TestInteg_ConsensusFailure(t *testing.T) {
	const script = `
		script {
			use 0x1::Account;
			use 0x1::DFI;
			
			fun main(account: &signer, recipient: address, amount: u128) {
				Account::pay_from_sender<DFI::T>(account, recipient, amount);
			}
		}
	`

	ct := cliTester.New(t, false)
	defer ct.Close()

	// Start DVM compiler container (runtime also, but we don't want for dnode to connect to DVM runtime)
	_, vmCompilerPort, err := server.FreeTCPAddr()
	require.NoError(t, err, "FreeTCPAddr for DVM compiler port")
	compilerStop := tests.LaunchDVMWithNetTransport(t, vmCompilerPort, ct.VMConnection.ListenPort, false)
	defer compilerStop()

	ct.SetVMCompilerAddressNet("tcp://127.0.0.1:"+vmCompilerPort, false)

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "script.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	// Should panic as there is no local VM running
	ct.TxVmExecuteScript(senderAddr, compiledPath, senderAddr, "100").DisableBroadcastMode().CheckSucceeded()

	// Check CONSENSUS FAILURE did occur
	{
		consensusFailure := false
		for i := 0; i < 10; i++ {
			if ct.DaemonLogsContain("CONSENSUS FAILURE") {
				consensusFailure = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		require.True(t, consensusFailure, "CONSENSUS FAILURE not occurred")
	}

	// Check restarted application panics
	{
		ct.RestartDaemon(false, false)

		retCode, daemonLogs := ct.CheckDaemonStopped(2 * time.Second)

		require.NotZero(t, retCode, "daemon exitCode")
		require.Contains(t, strings.Join(daemonLogs, ","), "panic", "daemon didn't panic")
	}
}

func TestIntegVM_ExecuteScript(t *testing.T) {
	const script = `
		script {
			use 0x1::Account;
			use 0x1::DFI;

			fun main(account: &signer) {
				let dfi = Account::withdraw_from_sender<DFI::T>(account, 1);
				Account::deposit_to_sender<DFI::T>(account, dfi);
			}
		}
	`

	ct := cliTester.New(
		t,
		true,
		cliTester.VMCommunicationOption(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()

	// Start DVM container
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "script.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "script.json")

	// Create .moe script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}

func TestIntegVM_RequestRetry(t *testing.T) {
	// TODO: Test should be rewritten as its success / failure is Moon phase dependant (not repeatable)
	t.Skip()

	const (
		dsSocket      = "ds.sock"
		mockDVMSocket = "mock_dvm.sock"
	)

	ct := cliTester.New(
		t,
		true,
		cliTester.VMCommunicationOption(100, 500, 10),
		cliTester.VMCommunicationBaseAddressUDSOption(dsSocket, mockDVMSocket),
	)
	defer ct.Close()
	ct.StartRestServer(false)

	mockDVMSocketPath := path.Join(ct.Dirs.UDSDir, mockDVMSocket)
	mockDVMListener, err := helpers.GetGRpcNetListener("unix://" + mockDVMSocketPath)
	require.NoError(t, err, "creating MockDVM listener")

	mockDvm := StartMockDVMService(mockDVMListener)
	defer mockDvm.Stop()
	require.NoError(t, testUtils.WaitForFileExists(mockDVMSocketPath, 1*time.Second), "MockDVM start failed")

	// Create fake .mov file
	modulePath := path.Join(ct.Dirs.RootDir, "fake.json")
	moduleContent := []byte("{ \"code\": \"00\" }")
	require.NoError(t, ioutil.WriteFile(modulePath, moduleContent, 0644), "creating fake module file")

	wg := sync.WaitGroup{}
	vmDeployDoneCh := make(chan bool)

	// Spam REST requests while dnode is stucked on VM request
	// Stop only once VM is done, that ensures routines were parallel
	{
		wg.Add(1)
		go func() {
			defer wg.Done()

			t.Logf("RestRequest: start")
			for {
				req, _ := ct.RestQueryOracleAssets()
				req.SetTimeout(1000 * time.Millisecond)
				req.CheckSucceeded()
				t.Logf("RestRequest: ok")

				select {
				case <-vmDeployDoneCh:
					t.Logf("RestRequest: stop")
					return
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	// Execute .json module file
	// That should take some time and when "done" we close the channel to stop the first routine
	{
		mockDvm.SetExecutionDelay(3 * time.Second)
		senderAddr := ct.Accounts["validator1"].Address

		wg.Add(1)
		go func() {
			defer func() {
				close(vmDeployDoneCh)
				wg.Done()
			}()

			t.Logf("VMDeploy: start")
			ct.TxVmDeployModule(senderAddr, modulePath).CheckSucceeded()
			t.Logf("VMDeploy: done")
		}()
	}

	wg.Wait()
}

// Test is skipped: should be used for dnode <-> dvm (uni-binary) communication over UDS debug locally (with DVM binaries).
func TestIntegVM_CommunicationUDS(t *testing.T) {
	t.Skip()

	const (
		dsSocket  = "ds.sock"
		dvmSocket = "dvm.sock"
	)

	const script = `
		script {
			use 0x1::Account;
			use 0x1::DFI;

			fun main(account: &signer) {
				let dfi = Account::withdraw_from_sender<DFI::T>(account, 1);
				Account::deposit_to_sender<DFI::T>(account, dfi);
			}
		}
	`

	t.Parallel()
	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressUDSOption(dsSocket, dvmSocket),
	)
	defer ct.Close()

	// Start DVM compiler / runtime (sub-process) abd register compiler
	os.Setenv(tests.EnvDvmIntegUse, "binary")
	dvmStop := tests.LaunchDVMWithUDSTransport(t, ct.Dirs.UDSDir, dvmSocket, dsSocket, false)
	defer dvmStop()

	ct.SetVMCompilerAddressUDS(path.Join(ct.Dirs.UDSDir, dvmSocket))

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "script.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}

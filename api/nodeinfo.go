package api

import (
	"net"
	"context"

	"github.com/bytom/version"
	"github.com/bytom/netsync"
	"github.com/bytom/errors"
	"github.com/bytom/p2p"
)

// NetInfo indicate net information
type NetInfo struct {
	Listening    bool   `json:"listening"`
	Syncing      bool   `json:"syncing"`
	Mining       bool   `json:"mining"`
	PeerCount    int    `json:"peer_count"`
	CurrentBlock uint64 `json:"current_block"`
	HighestBlock uint64 `json:"highest_block"`
	NetWorkID    string `json:"network_id"`
	Version      string `json:"version"`
}

// GetNodeInfo return net information
func (a *API) GetNodeInfo() *NetInfo {
	info := &NetInfo{
		Listening:    a.sync.Switch().IsListening(),
		Syncing:      a.sync.BlockKeeper().IsCaughtUp(),
		Mining:       a.cpuMiner.IsMining(),
		PeerCount:    len(a.sync.Switch().Peers().List()),
		CurrentBlock: a.chain.BestBlockHeight(),
		NetWorkID:    a.sync.NodeInfo().Network,
		Version:      version.Version,
	}
	_, info.HighestBlock = a.sync.Peers().BestPeer()
	if info.CurrentBlock > info.HighestBlock {
		info.HighestBlock = info.CurrentBlock
	}
	return info
}

// getPeerInfos return all peer information of current node
func (a *API) getPeerInfos() []*netsync.PeerInfo {
	peerSet := a.sync.Peers()
	peers := peerSet.Peers()

	var peerInfos []*netsync.PeerInfo

	for _, peer := range peers {
		peerInfos = append(peerInfos, peer.GetPeerInfo())
	}
	return peerInfos
}

// return the currently connected peers
func (a *API) connectedPeers() map[string]*netsync.PeerInfo {
	peerInfos := a.getPeerInfos()
	connectedPeers := make(map[string]*netsync.PeerInfo, len(peerInfos))
	for _, peerInfo := range peerInfos {
		connectedPeers[peerInfo.RemoteAddr] = peerInfo
	}
	return connectedPeers
}

// disconnect peer by the peer id
func (a *API) disconnectPeerById(peerId string) error {
	if peer, ok := a.sync.Peers().Peer(peerId); ok {
		swPeer := peer.GetPeer()
		a.sync.Switch().StopPeerGracefully(swPeer)
		return nil
	} else {
		return errors.New("peerId not exist")
	}
}

// connect peer b y net address
func (a *API) connectPeerByIpAndPort(ip string, port uint16) (*netsync.PeerInfo, error) {

	netIp := net.ParseIP(ip)
	if netIp == nil {
		return nil, errors.New("invalid ip address")
	}

	addr := p2p.NewNetAddressIPPort(netIp, port)
	sw := a.sync.Switch()

	if sw.NodeInfo().ListenAddr == addr.String() {
		return nil, errors.New("the dialing address is equals current node's address")
	}
	if dialling := sw.IsDialing(addr); dialling {
		return nil, errors.New("the address is dialing...")
	}
	if _, ok := a.connectedPeers()[addr.String()]; ok {
		return nil, errors.New("the address is already connected")
	}
	if err := sw.DialPeerWithAddress(addr); err != nil {
		return nil, errors.Wrap(err, "can not connect to the address")
	}
	return a.connectedPeers()[addr.String()], nil
}

// getNetInfo return network information
func (a *API) getNetInfo() Response {
	return NewSuccessResponse(a.GetNodeInfo())
}

// isMining return is in mining or not
func (a *API) isMining() Response {
	IsMining := map[string]bool{"is_mining": a.IsMining()}
	return NewSuccessResponse(IsMining)
}

// IsMining return mining status
func (a *API) IsMining() bool {
	return a.cpuMiner.IsMining()
}

// return the peers of current node
func (a *API) listPeers() Response {
	return NewSuccessResponse(a.getPeerInfos())
}

// disconnect peer
func (a *API) disconnectPeer(ctx context.Context, ins struct {
	PeerId string `json:"peerId"`
}) Response {
	if err := a.disconnectPeerById(ins.PeerId); err != nil {
		return NewErrorResponse(err)
	}
	return NewSuccessResponse(nil)
}

// connect peer by ip and port
func (a *API) connectPeer(ctx context.Context, ins struct {
	Ip   string `json:"ip"`
	Port uint16 `json:"port"`
}) Response {
	if peer, err := a.connectPeerByIpAndPort(ins.Ip, ins.Port); err != nil {
		return NewErrorResponse(err)
	} else {
		return NewSuccessResponse(peer)
	}
}

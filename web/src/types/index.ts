/**
 * Deployment configuration specific to Fabric peer nodes
 */
export interface FabricPeerDeploymentConfig {
  /** Organization ID that owns this peer */
  organizationId: number;
  /** MSP ID for the organization */
  mspId: string;
  /** ID of the signing key */
  signKeyId: number;
  /** ID of the TLS key */
  tlsKeyId: number;
  /** PEM encoded signing certificate */
  signCert: string;
  /** PEM encoded TLS certificate */
  tlsCert: string;
  /** PEM encoded CA certificate */
  caCert: string;
  /** PEM encoded TLS CA certificate */
  tlsCaCert: string;
  /** Listen address for the peer */
  listenAddress: string;
  /** Chaincode listen address */
  chaincodeAddress: string;
  /** Events listen address */
  eventsAddress: string;
  /** Operations listen address */
  operationsListenAddress: string;
  /** External endpoint for the peer */
  externalEndpoint: string;
  /** Domain names for the peer */
  domainNames?: string[];
}

export interface FabricOrdererDeploymentConfig {
  /** Organization ID that owns this orderer */
  organizationId: number;
  /** MSP ID for the organization */
  mspId: string;
  /** ID of the signing key */
  signKeyId: number;
  /** ID of the TLS key */
  tlsKeyId: number;
  /** PEM encoded signing certificate */
  signCert: string;
  /** PEM encoded TLS certificate */
  tlsCert: string;
  /** PEM encoded CA certificate */
  caCert: string;
  /** PEM encoded TLS CA certificate */
  tlsCaCert: string;
  /** Listen address for the orderer */
  listenAddress: string;
  /** Admin listen address */
  adminAddress: string;
  /** Operations listen address */
  operationsListenAddress: string;
  /** External endpoint for the orderer */
  externalEndpoint: string;
  /** Domain names for the orderer */
  domainNames?: string[];
}

/**
 * Deployment configuration specific to Besu nodes
 */
export interface BesuNodeDeploymentConfig {
  /** ID of the node key */
  keyId: number;
  /** P2P port for node communication */
  p2pPort: number;
  /** RPC port for API access */
  rpcPort: number;
  /** P2P host address */
  p2pHost: string;
  /** RPC host address */
  rpcHost: string;
  /** External IP address of the node */
  externalIp: string;
  /** Internal IP address of the node */
  internalIp: string;
  /** Network ID of the blockchain */
  networkId: number;
  /** Enode URL for node discovery */
  enodeUrl: string;
}

export type TabValue = 'details' | 'anchor-peers' | 'consenters'

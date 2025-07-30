package templates

const executorTemplateNew = `grpcPort: {{envDefault "EXECUTOR_PORT" "9090"}}
performerNetworkName: "{{envDefault "PERFORMER_NETWORK_NAME" "hgctl-performer-network"}}"

operator:
  address: "{{env "OPERATOR_ADDRESS"}}"
  {{if env "OPERATOR_PRIVATE_KEY"}}
  operatorPrivateKey:
    privateKey: "{{env "OPERATOR_PRIVATE_KEY"}}"
  {{end}}
  signingKeys:
    {{/* BLS Signer Configuration */}}
    {{with getSignerConfig "BLS"}}
    bls:
      {{if eq .Type "web3signer"}}
      remoteSigner: true
      remoteSignerConfig:
        url: "{{.Web3SignerURL}}"
        publicKey: "{{.Web3SignerPublicKey}}"
        {{if .Web3SignerCA}}
        caCert: |
{{.Web3SignerCA | indent 10}}
        {{else if env "WEB3_SIGNER_BLS_CA_CERT"}}
        caCert: "{{env "WEB3_SIGNER_BLS_CA_CERT"}}"
        {{end}}
        {{if .Web3SignerCert}}
        cert: |
{{.Web3SignerCert | indent 10}}
        {{else if env "WEB3_SIGNER_BLS_CLIENT_CERT"}}
        cert: "{{env "WEB3_SIGNER_BLS_CLIENT_CERT"}}"
        {{end}}
        {{if .Web3SignerKey}}
        key: |
{{.Web3SignerKey | indent 10}}
        {{else if env "WEB3_SIGNER_BLS_CLIENT_KEY"}}
        key: "{{env "WEB3_SIGNER_BLS_CLIENT_KEY"}}"
        {{end}}
      {{else if eq .Type "keystore"}}
      {{if .KeystoreContent}}
      keystore: |
{{.KeystoreContent | indent 8}}
      {{else if env "BLS_KEYSTORE_FILE"}}
      keystoreFile: "{{env "BLS_KEYSTORE_FILE"}}"
      {{else}}
      keystoreFile: "/keystores/operator.bls.keystore.json"
      {{end}}
      password: "{{or .KeystorePassword (env "BLS_KEYSTORE_PASSWORD") ""}}"
      {{else if eq .Type "privatekey"}}
      privateKey: "{{.PrivateKey}}"
      {{end}}
    {{end}}
    
    {{/* ECDSA Signer Configuration */}}
    {{with getSignerConfig "ECDSA"}}
    ecdsa:
      {{if eq .Type "web3signer"}}
      remoteSigner: true
      remoteSignerConfig:
        url: "{{.Web3SignerURL}}"
        publicKey: "{{.Web3SignerPublicKey}}"
        {{if .Web3SignerCA}}
        caCert: |
{{.Web3SignerCA | indent 10}}
        {{else if env "WEB3_SIGNER_ECDSA_CA_CERT"}}
        caCert: "{{env "WEB3_SIGNER_ECDSA_CA_CERT"}}"
        {{end}}
        {{if .Web3SignerCert}}
        cert: |
{{.Web3SignerCert | indent 10}}
        {{else if env "WEB3_SIGNER_ECDSA_CLIENT_CERT"}}
        cert: "{{env "WEB3_SIGNER_ECDSA_CLIENT_CERT"}}"
        {{end}}
        {{if .Web3SignerKey}}
        key: |
{{.Web3SignerKey | indent 10}}
        {{else if env "WEB3_SIGNER_ECDSA_CLIENT_KEY"}}
        key: "{{env "WEB3_SIGNER_ECDSA_CLIENT_KEY"}}"
        {{end}}
      {{else if eq .Type "keystore"}}
      {{if .KeystoreContent}}
      keystore: |
{{.KeystoreContent | indent 8}}
      {{else if env "ECDSA_KEYSTORE_FILE"}}
      keystoreFile: "{{env "ECDSA_KEYSTORE_FILE"}}"
      {{else}}
      keystoreFile: "/keystores/operator.ecdsa.keystore.json"
      {{end}}
      password: "{{or .KeystorePassword (env "ECDSA_KEYSTORE_PASSWORD") ""}}"
      {{else if eq .Type "privatekey"}}
      privateKey: "{{.PrivateKey}}"
      {{end}}
    {{end}}

l1Chain:
  chainId: "{{env "L1_CHAIN_ID"}}"
  rpcUrl: "{{env "L1_RPC_URL"}}"

avsPerformers:
  - image:
      repository: "{{env "PERFORMER_REGISTRY"}}"
      tag: "{{envDefault "PERFORMER_TAG" "latest"}}"
    processType: "{{envDefault "PERFORMER_PROCESS_TYPE" "server"}}"
    avsAddress: "{{env "AVS_ADDRESS"}}"
    workerCount: {{envDefault "WORKER_COUNT" "1"}}
    signingCurve: "{{envDefault "SIGNING_CURVE" "bn254"}}"
    {{if env "AVS_REGISTRAR_ADDRESS"}}
    avsRegistrarAddress: "{{env "AVS_REGISTRAR_ADDRESS"}}"
    {{end}}

{{if env "TASK_MAILBOX_L2_ADDRESS"}}
overrideContracts:
  taskMailbox:
    chainIds: [{{env "L1_CHAIN_ID"}}]
    contract: |
      {
        "name": "TaskMailbox",
        "address": "{{env "TASK_MAILBOX_L2_ADDRESS"}}",
        "chainId": {{envDefault "L2_CHAIN_ID" "8453"}}
      }
{{end}}`
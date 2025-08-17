package signerUtils

//func FromContext(ctx *config.Context, l logger.Logger) (signer.Signers, error) {
//	signers := signer.Signers{}
//	if ctx.SystemSignerKeys.BLS != nil {
//		var err error
//		var storedKeys *keystore.EIP2335Keystore
//		if ctx.SystemSignerKeys.BLS.Keystore != "" {
//			storedKeys, err = keystore.ParseKeystoreJSON(ctx.SystemSignerKeys.BLS.Keystore)
//			if err != nil {
//				return signers, fmt.Errorf("failed to parse keystore JSON: %w", err)
//			}
//		} else {
//			storedKeys, err = keystore.LoadKeystoreFile(ctx.SystemSignerKeys.BLS.KeystoreFile)
//			if err != nil {
//				return signers, fmt.Errorf("failed to load keystore file: '%s' %w", ctx.SystemSignerKeys.BLS.KeystoreFile, err)
//			}
//		}
//
//		privateSigningKey, err := storedKeys.GetBN254PrivateKey(ctx.SystemSignerKeys.BLS.Password)
//		if err != nil {
//			return signers, fmt.Errorf("failed to get private key: %w", err)
//		}
//
//		signers.BLSSigner = inMemorySigner.NewInMemorySigner(privateSigningKey, config.CurveTypeBN254)
//	}
//
//	if ctx.SystemSignerKeys.ECDSA != nil {
//		if ctx.SystemSignerKeys.ECDSA.UseRemoteSigner && ctx.SystemSignerKeys.ECDSA.RemoteSignerConfig != nil {
//			c, err := client.NewWeb3SignerClientFromRemoteSignerConfig(ctx.SystemSignerKeys.ECDSA.RemoteSignerConfig, l)
//			if err != nil {
//				return signers, fmt.Errorf("failed to create web3signer client: %w", err)
//			}
//			sig, err := web3Signer.NewWeb3Signer(
//				c,
//				common.HexToAddress(ctx.SystemSignerKeys.ECDSA.RemoteSignerConfig.FromAddress),
//				ctx.SystemSignerKeys.ECDSA.RemoteSignerConfig.PublicKey,
//				config.CurveTypeECDSA,
//				l,
//			)
//			if err != nil {
//				return signers, fmt.Errorf("failed to create web3 signer: %w", err)
//			}
//			signers.ECDSASigner = sig
//		} else if ctx.SystemSignerKeys.ECDSA.PrivateKey != "" {
//			ecdsaPk, err := ecdsa.NewPrivateKeyFromHexString(ctx.SystemSignerKeys.ECDSA.PrivateKey)
//			if err != nil {
//				return signers, fmt.Errorf("failed to create ECDSA private key: %w", err)
//			}
//			signers.ECDSASigner = inMemorySigner.NewInMemorySigner(ecdsaPk, config.CurveTypeECDSA)
//		} else {
//			l.Sugar().Warn("No ECDSA signing key provided")
//		}
//	}
//	return signers, nil
//}

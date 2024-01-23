# evmx
This module separates the EVM from [main ethereum implementation](https://github.com/ethereum/go-ethereum), still using some of shared libs from the main codebase. It's important to use the same version of go-ethereum as specified.

## Purpose

For research purposes, we needed to have a separate vm to be able to add it to our experimental blockchain, exploring the cross-chain world with IBC messages.

## References

Thanks to [go-evm](https://github.com/duanbing/go-evm) team, which we got the inspiration of how to do the separation from it.


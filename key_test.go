// Copyright (C) 2017. See AUTHORS.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openssl

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	pem_pkg "encoding/pem"
	"os"
	"testing"
)

func TestMarshal(t *testing.T) {
	key, err := LoadPrivateKeyFromPEM(keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := LoadCertificateFromPEM(certBytes)
	if err != nil {
		t.Fatal(err)
	}

	if !key.Equal(key) {
		t.Fatal("key not equal to itself")
	}

	privateBlock, _ := pem_pkg.Decode(keyBytes)
	key, err = LoadPrivateKeyFromDER(privateBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	pem, err := cert.MarshalPEM()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pem, certBytes) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", certBytes, 0644)
		t.Fatal("invalid cert pem bytes")
	}

	pem, err = key.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pem, keyBytes) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", keyBytes, 0644)
		t.Fatal("invalid private key pem bytes")
	}
	tls_cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	tls_key, ok := tls_cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		t.Fatal("FASDFASDF")
	}
	_ = tls_key

	der, err := key.MarshalPKCS1PrivateKeyDER()
	if err != nil {
		t.Fatal(err)
	}
	tls_der := x509.MarshalPKCS1PrivateKey(tls_key)
	if !bytes.Equal(der, tls_der) {
		t.Fatalf("invalid private key der bytes: %s\n v.s. %s\n",
			hex.Dump(der), hex.Dump(tls_der))
	}

	der, err = key.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}
	tls_der, err = x509.MarshalPKIXPublicKey(&tls_key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(der, tls_der) {
		os.WriteFile("generated", []byte(hex.Dump(der)), 0644)
		os.WriteFile("hardcoded", []byte(hex.Dump(tls_der)), 0644)
		t.Fatal("invalid public key der bytes")
	}

	pem, err = key.MarshalPKIXPublicKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	tls_pem := pem_pkg.EncodeToMemory(&pem_pkg.Block{
		Type: "PUBLIC KEY", Bytes: tls_der})
	if !bytes.Equal(pem, tls_pem) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", tls_pem, 0644)
		t.Fatal("invalid public key pem bytes")
	}

	loaded_pubkey_from_pem, err := LoadPublicKeyFromPEM(pem)
	if err != nil {
		t.Fatal(err)
	}

	loaded_pubkey_from_der, err := LoadPublicKeyFromDER(der)
	if err != nil {
		t.Fatal(err)
	}

	new_der_from_pem, err := loaded_pubkey_from_pem.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}

	new_der_from_der, err := loaded_pubkey_from_der.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(new_der_from_der, tls_der) {
		os.WriteFile("generated", []byte(hex.Dump(new_der_from_der)), 0644)
		os.WriteFile("hardcoded", []byte(hex.Dump(tls_der)), 0644)
		t.Fatal("invalid public key der bytes")
	}

	if !bytes.Equal(new_der_from_pem, tls_der) {
		os.WriteFile("generated", []byte(hex.Dump(new_der_from_pem)), 0644)
		os.WriteFile("hardcoded", []byte(hex.Dump(tls_der)), 0644)
		t.Fatal("invalid public key der bytes")
	}
}

func TestGenerate(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatal(err)
	}
	_, err = key.MarshalPKIXPublicKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	_, err = key.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateRSAKeyWithExponent(1024, 65537)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateEC(t *testing.T) {
	key, err := GenerateECKey(Prime256v1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = key.MarshalPKIXPublicKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	_, err = key.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateEd25519(t *testing.T) {
	if !ed25519_support {
		t.SkipNow()
	}

	key, err := GenerateED25519Key()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = key.MarshalPKIXPublicKeyPEM(); err != nil {
		t.Fatal(err)
	}
}

func TestSign(t *testing.T) {
	key, _ := GenerateRSAKey(1024)
	data := []byte("the quick brown fox jumps over the lazy dog")
	_, err := key.SignPKCS1v15(SHA1_Method, data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = key.SignPKCS1v15(SHA256_Method, data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = key.SignPKCS1v15(SHA512_Method, data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSignEC(t *testing.T) {
	t.Parallel()

	key, err := GenerateECKey(Prime256v1)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("the quick brown fox jumps over the lazy dog")

	t.Run("sha1", func(t *testing.T) {
		t.Parallel()
		sig, err := key.SignPKCS1v15(SHA1_Method, data)
		if err != nil {
			t.Fatal(err)
		}
		err = key.VerifyPKCS1v15(SHA1_Method, data, sig)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("sha256", func(t *testing.T) {
		t.Parallel()
		sig, err := key.SignPKCS1v15(SHA256_Method, data)
		if err != nil {
			t.Fatal(err)
		}
		err = key.VerifyPKCS1v15(SHA256_Method, data, sig)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("sha512", func(t *testing.T) {
		t.Parallel()
		sig, err := key.SignPKCS1v15(SHA512_Method, data)
		if err != nil {
			t.Fatal(err)
		}
		err = key.VerifyPKCS1v15(SHA512_Method, data, sig)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSignED25519(t *testing.T) {
	if !ed25519_support {
		t.SkipNow()
	}

	t.Parallel()

	key, err := GenerateED25519Key()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("the quick brown fox jumps over the lazy dog")

	t.Run("new", func(t *testing.T) {
		t.Parallel()
		sig, err := key.SignPKCS1v15(nil, data)
		if err != nil {
			t.Fatal(err)
		}

		err = key.VerifyPKCS1v15(nil, data, sig)
		if err != nil {
			t.Fatal(err)
		}

	})
}

func TestMarshalEC(t *testing.T) {
	if _, err := LoadPrivateKeyFromPEM(prime256v1KeyBytes); err != nil {
		t.Fatal(err)
	}
	cert, err := LoadCertificateFromPEM(prime256v1CertBytes)
	if err != nil {
		t.Fatal(err)
	}

	privateBlock, _ := pem_pkg.Decode(prime256v1KeyBytes)
	key, err := LoadPrivateKeyFromDER(privateBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	pem, err := cert.MarshalPEM()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pem, prime256v1CertBytes) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", prime256v1CertBytes, 0644)
		t.Fatal("invalid cert pem bytes")
	}

	pem, err = key.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pem, prime256v1KeyBytes) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", prime256v1KeyBytes, 0644)
		t.Fatal("invalid private key pem bytes")
	}
	tls_cert, err := tls.X509KeyPair(prime256v1CertBytes, prime256v1KeyBytes)
	if err != nil {
		t.Fatal(err)
	}
	tls_key, ok := tls_cert.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatal("FASDFASDF")
	}
	_ = tls_key

	der, err := key.MarshalPKCS1PrivateKeyDER()
	if err != nil {
		t.Fatal(err)
	}
	tls_der, err := x509.MarshalECPrivateKey(tls_key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(der, tls_der) {
		t.Fatalf("invalid private key der bytes: %s\n v.s. %s\n",
			hex.Dump(der), hex.Dump(tls_der))
	}

	der, err = key.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}
	tls_der, err = x509.MarshalPKIXPublicKey(&tls_key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(der, tls_der) {
		os.WriteFile("generated", []byte(hex.Dump(der)), 0644)
		os.WriteFile("hardcoded", []byte(hex.Dump(tls_der)), 0644)
		t.Fatal("invalid public key der bytes")
	}

	pem, err = key.MarshalPKIXPublicKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	tls_pem := pem_pkg.EncodeToMemory(&pem_pkg.Block{
		Type: "PUBLIC KEY", Bytes: tls_der})
	if !bytes.Equal(pem, tls_pem) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", tls_pem, 0644)
		t.Fatal("invalid public key pem bytes")
	}

	loaded_pubkey_from_pem, err := LoadPublicKeyFromPEM(pem)
	if err != nil {
		t.Fatal(err)
	}

	loaded_pubkey_from_der, err := LoadPublicKeyFromDER(der)
	if err != nil {
		t.Fatal(err)
	}

	new_der_from_pem, err := loaded_pubkey_from_pem.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}

	new_der_from_der, err := loaded_pubkey_from_der.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(new_der_from_der, tls_der) {
		os.WriteFile("generated", []byte(hex.Dump(new_der_from_der)), 0644)
		os.WriteFile("hardcoded", []byte(hex.Dump(tls_der)), 0644)
		t.Fatal("invalid public key der bytes")
	}

	if !bytes.Equal(new_der_from_pem, tls_der) {
		os.WriteFile("generated", []byte(hex.Dump(new_der_from_pem)), 0644)
		os.WriteFile("hardcoded", []byte(hex.Dump(tls_der)), 0644)
		t.Fatal("invalid public key der bytes")
	}
}

func TestMarshalEd25519(t *testing.T) {
	if !ed25519_support {
		t.SkipNow()
	}

	if _, err := LoadPrivateKeyFromPEM(ed25519KeyBytes); err != nil {
		t.Fatal(err)
	}
	cert, err := LoadCertificateFromPEM(ed25519CertBytes)
	if err != nil {
		t.Fatal(err)
	}

	privateBlock, _ := pem_pkg.Decode(ed25519KeyBytes)
	key, err := LoadPrivateKeyFromDER(privateBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	pem, err := cert.MarshalPEM()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pem, ed25519CertBytes) {
		os.WriteFile("generated", pem, 0644)
		os.WriteFile("hardcoded", ed25519CertBytes, 0644)
		t.Fatal("invalid cert pem bytes")
	}

	// NOTE: Ed25519 cannot be marshalled to PEM.

	if _, err := key.MarshalPKCS1PrivateKeyDER(); err != nil {
		t.Fatal(err)
	}

	der, err := key.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}

	pem, err = key.MarshalPKIXPublicKeyPEM()
	if err != nil {
		t.Fatal(err)
	}

	loadedPubkeyFromPem, err := LoadPublicKeyFromPEM(pem)
	if err != nil {
		t.Fatal(err)
	}

	loadedPubkeyFromDer, err := LoadPublicKeyFromDER(der)
	if err != nil {
		t.Fatal(err)
	}

	_, err = loadedPubkeyFromPem.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}

	_, err = loadedPubkeyFromDer.MarshalPKIXPublicKeyDER()
	if err != nil {
		t.Fatal(err)
	}
}

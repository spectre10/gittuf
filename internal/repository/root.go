// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gittuf/gittuf/internal/policy"
	"github.com/gittuf/gittuf/internal/signerverifier/dsse"
	sslibsv "github.com/gittuf/gittuf/internal/third_party/go-securesystemslib/signerverifier"
	"github.com/gittuf/gittuf/internal/tuf"
	sslibdsse "github.com/secure-systems-lab/go-securesystemslib/dsse"
)

// InitializeRoot is the interface for the user to create the repository's root
// of trust.
func (r *Repository) InitializeRoot(ctx context.Context, signer sslibdsse.SignerVerifier, signCommit bool) error {
	if err := r.InitializeNamespaces(); err != nil {
		return err
	}

	rawKey := signer.Public()
	publicKey, err := sslibsv.NewKey(rawKey)
	if err != nil {
		return err
	}

	rootMetadata := policy.InitializeRootMetadata(publicKey)

	env, err := dsse.CreateEnvelope(rootMetadata)
	if err != nil {
		return nil
	}

	env, err = dsse.SignEnvelope(ctx, env, signer)
	if err != nil {
		return nil
	}

	state := &policy.State{
		RootPublicKeys: []*tuf.Key{publicKey},
		RootEnvelope:   env,
	}

	commitMessage := "Initialize root of trust"

	return state.Commit(ctx, r.r, commitMessage, signCommit)
}

// AddRootKey is the interface for the user to add an authorized key
// for the Root role.
func (r *Repository) AddRootKey(ctx context.Context, signer sslibdsse.SignerVerifier, newRootKey *tuf.Key, signCommit bool) error {
	rootKeyID, err := signer.KeyID()
	if err != nil {
		return err
	}

	state, err := policy.LoadCurrentState(ctx, r.r)
	if err != nil {
		return err
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		return err
	}

	if !isKeyAuthorized(rootMetadata.Roles[policy.RootRoleName].KeyIDs, rootKeyID) {
		return ErrUnauthorizedKey
	}

	rootMetadata = policy.AddRootKey(rootMetadata, newRootKey)

	rootMetadata.SetVersion(rootMetadata.Version + 1)
	rootMetadataBytes, err := json.Marshal(rootMetadata)
	if err != nil {
		return err
	}

	env := state.RootEnvelope
	env.Signatures = []sslibdsse.Signature{}
	env.Payload = base64.StdEncoding.EncodeToString(rootMetadataBytes)

	env, err = dsse.SignEnvelope(ctx, env, signer)
	if err != nil {
		return err
	}

	state.RootEnvelope = env

	found := false
	for _, key := range state.RootPublicKeys {
		if key.KeyID == newRootKey.KeyID {
			found = true
			break
		}
	}
	if !found {
		state.RootPublicKeys = append(state.RootPublicKeys, newRootKey)
	}

	commitMessage := fmt.Sprintf("Add root key '%s' to root", newRootKey.KeyID)

	return state.Commit(ctx, r.r, commitMessage, signCommit)
}

// RemoveRootKey is the interface for the user to de-authorize a key
// trusted to sign the Root role.
func (r *Repository) RemoveRootKey(ctx context.Context, signer sslibdsse.SignerVerifier, keyID string, signCommit bool) error {
	rootKeyID, err := signer.KeyID()
	if err != nil {
		return err
	}

	state, err := policy.LoadCurrentState(ctx, r.r)
	if err != nil {
		return err
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		return err
	}

	if !isKeyAuthorized(rootMetadata.Roles[policy.RootRoleName].KeyIDs, rootKeyID) {
		return ErrUnauthorizedKey
	}

	rootMetadata, err = policy.DeleteRootKey(rootMetadata, keyID)
	if err != nil {
		return err
	}

	rootMetadata.SetVersion(rootMetadata.Version + 1)
	rootMetadataBytes, err := json.Marshal(rootMetadata)
	if err != nil {
		return err
	}

	env := state.RootEnvelope
	env.Signatures = []sslibdsse.Signature{}
	env.Payload = base64.StdEncoding.EncodeToString(rootMetadataBytes)

	env, err = dsse.SignEnvelope(ctx, env, signer)
	if err != nil {
		return err
	}

	newRootPublicKeys := []*tuf.Key{}
	for _, key := range state.RootPublicKeys {
		if key.KeyID != keyID {
			newRootPublicKeys = append(newRootPublicKeys, key)
		}
	}

	state.RootEnvelope = env
	state.RootPublicKeys = newRootPublicKeys

	commitMessage := fmt.Sprintf("Remove root key '%s' from root", keyID)

	return state.Commit(ctx, r.r, commitMessage, signCommit)
}

// AddTopLevelTargetsKey is the interface for the user to add an authorized key
// for the top level Targets role / policy file.
func (r *Repository) AddTopLevelTargetsKey(ctx context.Context, signer sslibdsse.SignerVerifier, targetsKey *tuf.Key, signCommit bool) error {
	rootKeyID, err := signer.KeyID()
	if err != nil {
		return err
	}

	state, err := policy.LoadCurrentState(ctx, r.r)
	if err != nil {
		return err
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		return err
	}

	if !isKeyAuthorized(rootMetadata.Roles[policy.RootRoleName].KeyIDs, rootKeyID) {
		return ErrUnauthorizedKey
	}

	rootMetadata = policy.AddTargetsKey(rootMetadata, targetsKey)

	rootMetadata.SetVersion(rootMetadata.Version + 1)
	rootMetadataBytes, err := json.Marshal(rootMetadata)
	if err != nil {
		return err
	}

	env := state.RootEnvelope
	env.Signatures = []sslibdsse.Signature{}
	env.Payload = base64.StdEncoding.EncodeToString(rootMetadataBytes)

	env, err = dsse.SignEnvelope(ctx, env, signer)
	if err != nil {
		return err
	}

	state.RootEnvelope = env

	commitMessage := fmt.Sprintf("Add policy key '%s' to root", targetsKey.KeyID)

	return state.Commit(ctx, r.r, commitMessage, signCommit)
}

// RemoveTopLevelTargetsKey is the interface for the user to de-authorize a key
// trusted to sign the top level Targets role / policy file.
func (r *Repository) RemoveTopLevelTargetsKey(ctx context.Context, signer sslibdsse.SignerVerifier, targetsKeyID string, signCommit bool) error {
	rootKeyID, err := signer.KeyID()
	if err != nil {
		return err
	}

	state, err := policy.LoadCurrentState(ctx, r.r)
	if err != nil {
		return err
	}

	rootMetadata, err := state.GetRootMetadata()
	if err != nil {
		return err
	}

	if !isKeyAuthorized(rootMetadata.Roles[policy.RootRoleName].KeyIDs, rootKeyID) {
		return ErrUnauthorizedKey
	}

	rootMetadata, err = policy.DeleteTargetsKey(rootMetadata, targetsKeyID)
	if err != nil {
		return err
	}

	rootMetadata.SetVersion(rootMetadata.Version + 1)
	rootMetadataBytes, err := json.Marshal(rootMetadata)
	if err != nil {
		return err
	}

	env := state.RootEnvelope
	env.Signatures = []sslibdsse.Signature{}
	env.Payload = base64.StdEncoding.EncodeToString(rootMetadataBytes)

	env, err = dsse.SignEnvelope(ctx, env, signer)
	if err != nil {
		return err
	}

	state.RootEnvelope = env

	commitMessage := fmt.Sprintf("Remove policy key '%s' from root", targetsKeyID)

	return state.Commit(ctx, r.r, commitMessage, signCommit)
}

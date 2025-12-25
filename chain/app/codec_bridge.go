// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var (
	registeredFiles int
	failedFiles     []string
	registeredTypes int
	failedTypes     []string
)

// init bridges protobuf file descriptors registered by the modern protobuf
// runtime into the legacy gogo registry. The Cosmos SDK's msgservice helper
// relies on gogoproto.RegisterFile when registering Msg service descriptors,
// so without this bridge ValidateGenesis and CLI plumbing panic when the
// descriptors are missing.
func init() {
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		path := fd.Path()
		if !strings.HasPrefix(path, "aura/") {
			return true
		}

		// Skip files already registered with gogo to avoid duplicate panics.
		if gogoproto.FileDescriptor(path) != nil {
			return true
		}

		fdProto := protodesc.ToFileDescriptorProto(fd)
		if fdProto == nil {
			return true
		}

		raw, err := proto2.Marshal(fdProto)
		if err != nil {
			return true
		}

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(raw); err == nil {
			_ = gz.Close()
			// Register the gzipped descriptor so msgservice.RegisterMsgServiceDesc
			// can look it up via proto.FileDescriptor(metadata).
			func() {
				defer func() {
					if r := recover(); r != nil {
						failedFiles = append(failedFiles, fmt.Sprintf("%s: %v", path, r))
					}
				}()
				gogoproto.RegisterFile(path, buf.Bytes())
				registeredFiles++
			}()
		}

		return true
	})

	// Also register message types with the legacy gogo registry so
	// msgservice.RegisterMsgServiceDesc can resolve requests/responses.
	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		fullName := string(mt.Descriptor().FullName())
		if !strings.HasPrefix(fullName, "aura.") {
			return true
		}

		msg := mt.Zero().Interface()
		legacyMsg, ok := msg.(interface {
			ProtoMessage()
			Reset()
			String() string
		})
		if !ok {
			return true
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					failedTypes = append(failedTypes, fmt.Sprintf("%s: %v", fullName, r))
				}
			}()
			gogoproto.RegisterType(legacyMsg, fullName)
			registeredTypes++
		}()
		return true
	})

	// Log summary at startup
	if len(failedFiles) > 0 || len(failedTypes) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: codec_bridge registration results:\n")
		fmt.Fprintf(os.Stderr, "  Files: %d registered, %d failed\n", registeredFiles, len(failedFiles))
		fmt.Fprintf(os.Stderr, "  Types: %d registered, %d failed\n", registeredTypes, len(failedTypes))

		if len(failedFiles) > 0 {
			fmt.Fprintf(os.Stderr, "Failed file registrations:\n")
			for _, f := range failedFiles {
				fmt.Fprintf(os.Stderr, "  - %s\n", f)
			}
		}

		if len(failedTypes) > 0 {
			fmt.Fprintf(os.Stderr, "Failed type registrations:\n")
			for _, t := range failedTypes {
				fmt.Fprintf(os.Stderr, "  - %s\n", t)
			}
		}
	}
}

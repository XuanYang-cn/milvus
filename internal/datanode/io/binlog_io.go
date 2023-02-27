// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io

import (
	"context"

	"github.com/milvus-io/milvus/internal/datanode/resource"
	"github.com/milvus-io/milvus/internal/storage"

	"github.com/milvus-io/milvus/pkg/util/retry"
)

type BinlogIO struct {
	storage.ChunkManager
	resources *resource.GlobalResources
}

func NewBinlogIO(cm storage.ChunkManager, resources *resource.GlobalResources) *BinlogIO {
	return &BinlogIO{cm, resources}
}

func (b *BinlogIO) Download(ctx context.Context, paths []string) ([][]byte, error) {
	var (
		vs  [][]byte
		err error
	)

	retry.Do(ctx, func() error {
		b.resources.SubmitIO(len(paths))
		vs, err = b.MultiRead(ctx, paths)
		return err
	})

	if err != nil {
		return nil, err
	}

	return vs, nil
}

func (b *BinlogIO) Upload(ctx context.Context, kvs map[string][]byte) error {
	var err error

	retry.Do(ctx, func() error {
		b.resources.SubmitIO(len(kvs))
		return b.MultiWrite(ctx, kvs)
	})

	return err
}

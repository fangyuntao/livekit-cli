// Copyright 2024 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"os"
	"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const flagRequest = "request"

type protoType[T any] interface {
	*T
	proto.Message
}

func ReadRequest[T any, P protoType[T]](c *cli.Context) (*T, error) {
	return ReadRequestFile[T, P](c.String(flagRequest))
}

func ReadRequestFile[T any, P protoType[T]](path string) (*T, error) {
	reqBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var req P = new(T)
	err = protojson.Unmarshal(reqBytes, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func RequestFlag[T any, P protoType[T]]() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     flagRequest,
		Usage:    RequestDesc[T, P](),
		Required: true,
	}
}

func RequestDesc[T any, _ protoType[T]]() string {
	typ := reflect.TypeFor[T]().Name()
	return typ + " as JSON file (see livekit-cli/examples)"
}

func createAndPrint[T any, P protoType[T], R any](
	c *cli.Context, file string,
	create func(ctx context.Context, p P) (R, error),
	print func(r R),
) error {
	req, err := ReadRequestFile[T, P](file)
	if err != nil {
		return err
	}
	if c.Bool("verbose") {
		PrintJSON(req)
	}
	info, err := create(c.Context, req)
	if err != nil {
		return err
	}
	print(info)
	return nil
}

func createAndPrintLegacy[T any, P protoType[T], R any](
	c *cli.Context,
	create func(ctx context.Context, p P) (R, error),
	print func(r R),
) error {
	req, err := ReadRequest[T, P](c)
	if err != nil {
		return err
	}
	if c.Bool("verbose") {
		PrintJSON(req)
	}
	info, err := create(c.Context, req)
	if err != nil {
		return err
	}
	print(info)
	return nil
}

func createAndPrintReqs[T any, P protoType[T], R any](
	c *cli.Context,
	create func(ctx context.Context, p P) (R, error),
	print func(r R),
) error {
	args := c.Args()
	if !args.Present() {
		return errors.New("at least one JSON request file is required")
	}
	for _, file := range args.Slice() {
		if err := createAndPrint(c, file, create, print); err != nil {
			return err
		}
	}
	return nil
}

func forEachID(c *cli.Context, fnc func(ctx context.Context, id string) error) error {
	args := c.Args()
	if !args.Present() {
		return errors.New("at least one ID is required")
	}
	for _, id := range args.Slice() {
		if err := fnc(c.Context, id); err != nil {
			return err
		}
	}
	return nil
}

func listAndPrint[
	ReqT any, Req protoType[ReqT],
	T any, _ protoType[T],
	Resp interface {
		proto.Message
		GetItems() []*T
	},
](
	c *cli.Context,
	getList func(ctx context.Context, req Req) (Resp, error), req Req,
	header []string, tableRow func(item *T) []string,
) error {
	res, err := getList(c.Context, req)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	for _, item := range res.GetItems() {
		if item == nil {
			continue
		}
		row := tableRow(item)
		if len(row) == 0 {
			continue
		}
		table.Append(row)
	}
	table.Render()
	if c.Bool("verbose") {
		PrintJSON(res)
	}
	return nil
}

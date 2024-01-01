// Copyright 2022 Praetorian Security, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linuxrpc

import (
    "testing"

    "github.com/ory/dockertest/v3"
    "github.com/yhy0/Jie/lib/fingerprintx/pkg/plugins"
    "github.com/yhy0/Jie/lib/fingerprintx/pkg/test"
)

func TestRPC(t *testing.T) {
    testcases := []test.Testcase{
        {
            Description: "alpine-nfs",
            Port:        111,
            Protocol:    plugins.TCP,
            Expected: func(res *plugins.Service) bool {
                return res != nil
            },
            RunConfig: dockertest.RunOptions{
                Repository: "woahbase/alpine-nfs",
                Tag:        "x86_64",
                Privileged: true,
            },
        },
    }

    p := &RPCPlugin{}

    for _, tc := range testcases {
        tc := tc
        t.Run(tc.Description, func(t *testing.T) {
            t.Parallel()
            err := test.RunTest(t, tc, p)
            if err != nil {
                t.Errorf(err.Error())
            }
        })
    }
}

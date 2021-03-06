/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1_test

import (
	"encoding/json"
	"reflect"
	"testing"

	newer "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/resource"
	current "github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1beta1"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
)

var Convert = newer.Scheme.Convert

func TestNodeConversion(t *testing.T) {
	version, kind, err := newer.Scheme.ObjectVersionAndKind(&current.Minion{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "v1beta1" || kind != "Minion" {
		t.Errorf("unexpected version and kind: %s %s", version, kind)
	}

	newer.Scheme.Log(t)
	obj, err := current.Codec.Decode([]byte(`{"kind":"Node","apiVersion":"v1beta1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := obj.(*newer.Node); !ok {
		t.Errorf("unexpected type: %#v", obj)
	}

	obj, err = current.Codec.Decode([]byte(`{"kind":"NodeList","apiVersion":"v1beta1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := obj.(*newer.NodeList); !ok {
		t.Errorf("unexpected type: %#v", obj)
	}

	obj = &newer.Node{}
	if err := current.Codec.DecodeInto([]byte(`{"kind":"Node","apiVersion":"v1beta1"}`), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj = &newer.Node{}
	data, err := current.Codec.Encode(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["kind"] != "Minion" {
		t.Errorf("unexpected encoding: %s - %#v", m["kind"], string(data))
	}
}

func TestEnvConversion(t *testing.T) {
	nonCanonical := []current.EnvVar{
		{Key: "EV"},
		{Key: "EV", Name: "EX"},
	}
	canonical := []newer.EnvVar{
		{Name: "EV"},
		{Name: "EX"},
	}
	for i := range nonCanonical {
		var got newer.EnvVar
		err := Convert(&nonCanonical[i], &got)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e, a := canonical[i], got; !reflect.DeepEqual(e, a) {
			t.Errorf("expected %v, got %v", e, a)
		}
	}

	// Test conversion the other way, too.
	for i := range canonical {
		var got current.EnvVar
		err := Convert(&canonical[i], &got)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e, a := canonical[i].Name, got.Key; e != a {
			t.Errorf("expected %v, got %v", e, a)
		}
		if e, a := canonical[i].Name, got.Name; e != a {
			t.Errorf("expected %v, got %v", e, a)
		}
	}
}

func TestVolumeMountConversionToOld(t *testing.T) {
	table := []struct {
		in  newer.VolumeMount
		out current.VolumeMount
	}{
		{
			in:  newer.VolumeMount{Name: "foo", MountPath: "/dev/foo", ReadOnly: true},
			out: current.VolumeMount{Name: "foo", MountPath: "/dev/foo", Path: "/dev/foo", ReadOnly: true},
		},
	}
	for _, item := range table {
		got := current.VolumeMount{}
		err := Convert(&item.in, &got)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		if e, a := item.out, got; !reflect.DeepEqual(e, a) {
			t.Errorf("Expected: %#v, got %#v", e, a)
		}
	}
}

func TestVolumeMountConversionToNew(t *testing.T) {
	table := []struct {
		in  current.VolumeMount
		out newer.VolumeMount
	}{
		{
			in:  current.VolumeMount{Name: "foo", MountPath: "/dev/foo", ReadOnly: true},
			out: newer.VolumeMount{Name: "foo", MountPath: "/dev/foo", ReadOnly: true},
		}, {
			in:  current.VolumeMount{Name: "foo", MountPath: "/dev/foo", Path: "/dev/bar", ReadOnly: true},
			out: newer.VolumeMount{Name: "foo", MountPath: "/dev/foo", ReadOnly: true},
		}, {
			in:  current.VolumeMount{Name: "foo", Path: "/dev/bar", ReadOnly: true},
			out: newer.VolumeMount{Name: "foo", MountPath: "/dev/bar", ReadOnly: true},
		},
	}
	for _, item := range table {
		got := newer.VolumeMount{}
		err := Convert(&item.in, &got)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		if e, a := item.out, got; !reflect.DeepEqual(e, a) {
			t.Errorf("Expected: %#v, got %#v", e, a)
		}
	}
}

func TestMinionListConversionToNew(t *testing.T) {
	oldMinion := func(id string) current.Minion {
		return current.Minion{TypeMeta: current.TypeMeta{ID: id}}
	}
	newNode := func(id string) newer.Node {
		return newer.Node{ObjectMeta: newer.ObjectMeta{Name: id}}
	}
	oldMinions := []current.Minion{
		oldMinion("foo"),
		oldMinion("bar"),
	}
	newMinions := []newer.Node{
		newNode("foo"),
		newNode("bar"),
	}

	table := []struct {
		oldML *current.MinionList
		newML *newer.NodeList
	}{
		{
			oldML: &current.MinionList{Items: oldMinions},
			newML: &newer.NodeList{Items: newMinions},
		}, {
			oldML: &current.MinionList{Minions: oldMinions},
			newML: &newer.NodeList{Items: newMinions},
		}, {
			oldML: &current.MinionList{
				Items:   oldMinions,
				Minions: []current.Minion{oldMinion("baz")},
			},
			newML: &newer.NodeList{Items: newMinions},
		},
	}

	for _, item := range table {
		got := &newer.NodeList{}
		err := Convert(item.oldML, got)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if e, a := item.newML, got; !newer.Semantic.DeepEqual(e, a) {
			t.Errorf("Expected: %#v, got %#v", e, a)
		}
	}
}

func TestMinionListConversionToOld(t *testing.T) {
	oldMinion := func(id string) current.Minion {
		return current.Minion{TypeMeta: current.TypeMeta{ID: id}}
	}
	newNode := func(id string) newer.Node {
		return newer.Node{ObjectMeta: newer.ObjectMeta{Name: id}}
	}
	oldMinions := []current.Minion{
		oldMinion("foo"),
		oldMinion("bar"),
	}
	newMinions := []newer.Node{
		newNode("foo"),
		newNode("bar"),
	}

	newML := &newer.NodeList{Items: newMinions}
	oldML := &current.MinionList{
		Items:   oldMinions,
		Minions: oldMinions,
	}

	got := &current.MinionList{}
	err := Convert(newML, got)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if e, a := oldML, got; !newer.Semantic.DeepEqual(e, a) {
		t.Errorf("Expected: %#v, got %#v", e, a)
	}
}

func TestServiceEmptySelector(t *testing.T) {
	// Nil map should be preserved
	svc := &current.Service{Selector: nil}
	data, err := newer.Scheme.EncodeToVersion(svc, "v1beta1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj, err := newer.Scheme.Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	selector := obj.(*newer.Service).Spec.Selector
	if selector != nil {
		t.Errorf("unexpected selector: %#v", obj)
	}

	// Empty map should be preserved
	svc2 := &current.Service{Selector: map[string]string{}}
	data, err = newer.Scheme.EncodeToVersion(svc2, "v1beta1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj, err = newer.Scheme.Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	selector = obj.(*newer.Service).Spec.Selector
	if selector == nil || len(selector) != 0 {
		t.Errorf("unexpected selector: %#v", obj)
	}
}

func TestPullPolicyConversion(t *testing.T) {
	table := []struct {
		versioned current.PullPolicy
		internal  newer.PullPolicy
	}{
		{
			versioned: current.PullAlways,
			internal:  newer.PullAlways,
		}, {
			versioned: current.PullNever,
			internal:  newer.PullNever,
		}, {
			versioned: current.PullIfNotPresent,
			internal:  newer.PullIfNotPresent,
		}, {
			versioned: "",
			internal:  "",
		}, {
			versioned: "invalid value",
			internal:  "invalid value",
		},
	}
	for _, item := range table {
		var got newer.PullPolicy
		err := Convert(&item.versioned, &got)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		if e, a := item.internal, got; e != a {
			t.Errorf("Expected: %q, got %q", e, a)
		}
	}
	for _, item := range table {
		var got current.PullPolicy
		err := Convert(&item.internal, &got)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		if e, a := item.versioned, got; e != a {
			t.Errorf("Expected: %q, got %q", e, a)
		}
	}
}

func getResourceRequirements(cpu, memory resource.Quantity) current.ResourceRequirementSpec {
	res := current.ResourceRequirementSpec{}
	res.Limits = current.ResourceList{}
	if cpu.Value() > 0 {
		res.Limits[current.ResourceCPU] = util.NewIntOrStringFromInt(int(cpu.Value()))
	}
	if memory.Value() > 0 {
		res.Limits[current.ResourceMemory] = util.NewIntOrStringFromInt(int(memory.Value()))
	}

	return res
}

func TestContainerConversion(t *testing.T) {
	cpuLimit := resource.MustParse("10")
	memoryLimit := resource.MustParse("10M")
	null := resource.Quantity{}
	testCases := []current.Container{
		{
			Name:      "container",
			Resources: getResourceRequirements(cpuLimit, memoryLimit),
		},
		{
			Name:      "container",
			CPU:       int(cpuLimit.MilliValue()),
			Resources: getResourceRequirements(null, memoryLimit),
		},
		{
			Name:      "container",
			Memory:    memoryLimit.Value(),
			Resources: getResourceRequirements(cpuLimit, null),
		},
		{
			Name:   "container",
			CPU:    int(cpuLimit.MilliValue()),
			Memory: memoryLimit.Value(),
		},
		{
			Name:      "container",
			Memory:    memoryLimit.Value(),
			Resources: getResourceRequirements(cpuLimit, resource.MustParse("100M")),
		},
		{
			Name:      "container",
			CPU:       int(cpuLimit.MilliValue()),
			Resources: getResourceRequirements(resource.MustParse("500"), memoryLimit),
		},
	}

	for i, tc := range testCases {
		got := newer.Container{}
		if err := Convert(&tc, &got); err != nil {
			t.Errorf("[Case: %d] Unexpected error: %v", i, err)
			continue
		}
		if cpu := got.Resources.Limits.Cpu(); cpu.Value() != cpuLimit.Value() {
			t.Errorf("[Case: %d] Expected cpu: %v, got: %v", i, cpuLimit, *cpu)
		}
		if memory := got.Resources.Limits.Memory(); memory.Value() != memoryLimit.Value() {
			t.Errorf("[Case: %d] Expected memory: %v, got: %v", i, memoryLimit, *memory)
		}
	}
}

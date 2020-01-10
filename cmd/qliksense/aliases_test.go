package main

import (
	"reflect"
	"testing"
)

func Test_stripFlagFromArgs(t *testing.T) {
	type args struct {
		args     []string
		flagName string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "valid case",
			
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripFlagFromArgs(tt.args.args, tt.args.flagName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripFlagFromArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

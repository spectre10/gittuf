package logging

import (
	"flag"

	"sync/atomic"

	"k8s.io/klog/v2"
)

var isInitialized atomic.Bool

func Print(msg string) {
	defer klog.Flush()
	if isInitialized.Load() {
		if klog.V(1).Enabled() {
			// log to both stderr and log_file
			klog.Warning("=> " + msg)
		} else {
			// only logs to log_file
			klog.Info("=> " + msg)
		}
	} else {
		klog.Error("Logger is not initialized!")
	}
}

func setFlags(fs *flag.FlagSet) error {
	err := fs.Set("logtostderr", "false")
	if err != nil {
		return err
	}

	// log to both stderr and log_file if klog.Warning or above called
	err = fs.Set("stderrthreshold", "1")
	if err != nil {
		return err
	}

	err = fs.Set("skip_headers", "true")
	if err != nil {
		return err
	}

	err = fs.Set("log_file", "./infotest.log")
	if err != nil {
		return err
	}
	return nil
}

func InitLogger(isVerbose bool) error {
	var fs flag.FlagSet
	klog.InitFlags(&fs)

	err := setFlags(&fs)
	if err != nil {
		return err
	}

	// set verbosity level
	if isVerbose {
		err = fs.Set("v", "1")
	} else {
		err = fs.Set("v", "0")
	}

	if err != nil {
		return err
	}

	isInitialized.Store(true)
	return nil
}

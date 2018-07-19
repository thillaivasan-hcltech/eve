// Copyright (c) 2017-2018 Zededa, Inc.
// All rights reserved.

package zedrouter

import (
	"github.com/zededa/go-provision/cast"
	"github.com/zededa/go-provision/devicenetwork"
	"github.com/zededa/go-provision/types"
	"log"
	"reflect"
)

func handleDNCModify(ctxArg interface{}, key string, configArg interface{}) {

	config := cast.CastDeviceNetworkConfig(configArg)
	ctx := ctxArg.(*zedrouterContext)

	if key != ctx.manufacturerModel {
		if debug {
			log.Printf("handleDNCModify: ignoring %s - expecting %s\n",
				key, ctx.manufacturerModel)
		}
		return
	}
	log.Printf("handleDNCModify for %s\n", key)

	deviceNetworkConfig = config
	new, _ := devicenetwork.MakeDeviceNetworkStatus(config,
		deviceNetworkStatus)
	// XXX switch to Equal?
	if !reflect.DeepEqual(deviceNetworkStatus, new) {
		log.Printf("DeviceNetworkStatus change from %v to %v\n",
			deviceNetworkStatus, new)
		deviceNetworkStatus = new
		doDNSUpdate(ctx)
	}
	log.Printf("handleDNCModify done for %s\n", key)
}

func handleDNCDelete(ctxArg interface{}, key string, configArg interface{}) {

	log.Printf("handleDNCDelete for %s\n", key)
	ctx := ctxArg.(*zedrouterContext)

	if key != "global" {
		log.Printf("handleDNCDelete: ignoring %s\n", key)
		return
	}
	new := types.DeviceNetworkStatus{}
	// XXX switch to Equal?
	if !reflect.DeepEqual(deviceNetworkStatus, new) {
		log.Printf("DeviceNetworkStatus change from %v to %v\n",
			deviceNetworkStatus, new)
		deviceNetworkStatus = new
		doDNSUpdate(ctx)
	}
	log.Printf("handleDNCDelete done for %s\n", key)
}

func doDNSUpdate(ctx *zedrouterContext) {
	// Did we loose all usable addresses or gain the first usable
	// address?
	newAddrCount := types.CountLocalAddrAnyNoLinkLocal(deviceNetworkStatus)
	if newAddrCount == 0 && ctx.usableAddressCount != 0 {
		log.Printf("DeviceNetworkStatus from %d to %d addresses\n",
			ctx.usableAddressCount, newAddrCount)
		// Inform ledmanager that we have no addresses
		types.UpdateLedManagerConfig(1)
	} else if newAddrCount != 0 && ctx.usableAddressCount == 0 {
		log.Printf("DeviceNetworkStatus from %d to %d addresses\n",
			ctx.usableAddressCount, newAddrCount)
		// Inform ledmanager that we have uplink addresses
		types.UpdateLedManagerConfig(2)
	}
	ctx.usableAddressCount = newAddrCount
	if !ctx.ready {
		return
	}
	publishDeviceNetworkStatus(ctx.pubDeviceNetworkStatus)
	updateLispConfiglets(ctx, ctx.separateDataPlane)
	setUplinks(deviceNetworkConfig.Uplink)
	setFreeUplinks(deviceNetworkConfig.FreeUplinks)
	// XXX do a NatInactivate/NatActivate if freeuplinks/uplinks changed?
}

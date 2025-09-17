package main

import (
	"fmt"
	u "pnas/internal/utils"
	"testing"
)

func TestLVMGetPVS(t *testing.T) {
	lvm := &u.LVM{}

	fmt.Println("==========PV==========")
	result, _ := lvm.GetPVS()
	for _, pv := range result {
		fmt.Println(pv)
	}

	fmt.Println("==========VG==========")
	// 获取卷组信息
	vgResult, _ := lvm.GetVGS()
	for _, vg := range vgResult {
		fmt.Println(vg)
	}

	fmt.Println("==========LV==========")
	// 获取逻辑卷信息
	lvResult, err := lvm.GetLVS()
	if err != nil {
		t.Errorf("获取逻辑卷信息失败：%v", err)
	}
	for _, lv := range lvResult {
		fmt.Println(lv)
	}

	// fmt.Println("==========PVCreate==========")
	// if err := lvm.CreatePV([]string{"/dev/sdc", "/dev/sdd"}); err != nil {
	// 	t.Errorf("创建物理卷失败：%v", err)
	// }
	//
	// fmt.Println("==========PVRemove==========")
	// if err := lvm.RemovePV([]string{"/dev/sdc"}); err != nil {
	// 	t.Errorf("删除物理卷失败：%v", err)
	// }
	//
	// fmt.Println("==========VGCreate==========")
	// if err := lvm.CreateVG("myvg", []string{"/dev/sdb"}); err != nil {
	// 	t.Errorf("创建卷组失败：%v", err)
	// }
	//
	// fmt.Println("==========VGRemove==========")
	// if err := lvm.RemoveVG("myvg"); err != nil {
	// 	t.Errorf("删除卷组失败：%v", err)
	// }
	fmt.Println("==========LVCreate==========")
	if err := lvm.CreateLV("myvg", "mylv", "9G"); err != nil {
		t.Errorf("创建逻辑卷失败：%v", err)
	}

	fmt.Println("==========LVRemove==========")
	if err := lvm.RemoveLV("myvg", "mylv"); err != nil {
		t.Errorf("删除逻辑卷失败：%v", err)
	}
}

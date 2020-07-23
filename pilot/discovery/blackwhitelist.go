package discovery

import (
	"fmt"
	"regexp"
	"strings"
)

type BlackWhiteList struct {
	bListNS  string
	wListNS  string
	bListPod string
	wListPod string
}

func (b *BlackWhiteList) IsResponsible(namespace string, pod string) (bool, error) {
	bListPodRegex, err := regexp.Compile(b.bListPod)
	if err != nil {
		return false, err
	}
	wListSVCRegex, err := regexp.Compile(b.wListPod)
	if err != nil {
		return false, err
	}
	podFilterKey := fmt.Sprintf("%v/%v", namespace, pod)
	if len(bListPodRegex.String()) > 0 {
		if match := bListPodRegex.MatchString(podFilterKey); match {
			return false, nil
		}
	}
	if len(wListSVCRegex.String()) > 0 {
		match := wListSVCRegex.MatchString(podFilterKey)
		return match || b.namespaceFilter(namespace), nil
	}
	return b.namespaceFilter(namespace), err
}
func (b *BlackWhiteList) namespaceFilter(namespace string) bool {
	bListNS := listToSet(parseList(b.bListNS))
	wListNS := listToSet(parseList(b.wListNS))
	if _, inBList := bListNS[namespace]; inBList {
		return false
	}
	if len(b.wListNS) > 0 {
		_, inWList := wListNS[namespace]
		return inWList
	}
	return true
}
func parseList(raw string) []string {
	if raw == "" {
		return nil
	}
	splitted := strings.Split(raw, ",")
	for i := range splitted {
		splitted[i] = strings.Trim(splitted[i], " \n\t")
	}
	return splitted
}

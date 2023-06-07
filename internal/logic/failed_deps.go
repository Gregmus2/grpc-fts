package logic

type failedDependencies map[string]struct{}

func (d failedDependencies) HasDependencyFailed(dependsOn []string) (bool, string) {
	for _, dependency := range dependsOn {
		if _, ok := d[dependency]; ok {
			return true, dependency
		}
	}

	return false, ""
}

func (d failedDependencies) Add(name string) {
	d[name] = struct{}{}
}

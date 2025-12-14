package cli

type commandSchedule struct {
	set    commandScheduleSet
	list   commandScheduleList
	remove commandScheduleRemove
	status commandScheduleStatus
}

func (c *commandSchedule) setup(svc advancedAppServices, parent commandParent) {
	cmd := parent.Command("schedule", "Manage OS-native scheduled backups")

	c.set.setup(svc, cmd)
	c.list.setup(svc, cmd)
	c.remove.setup(svc, cmd)
	c.status.setup(svc, cmd)
}

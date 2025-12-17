package domain

type MachineState string

const MachineStateNotExist = MachineState("notexist")
const MachineStateRunning = MachineState("running")
const MachineStateStopped = MachineState("stopped")

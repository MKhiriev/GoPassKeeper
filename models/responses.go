package models

type SyncResponse struct {
	PrivateDataStates []PrivateDataState `json:"private_data_states"`
	Length            int                `json:"length"`
}

package service

type ReceiveAddressResponse struct {
	LiquidAddress         string `binding:"required" json:"liquid-address"`
	PlanetmintBeneficiary string `binding:"required" json:"planetmint-beneficiary"`
}

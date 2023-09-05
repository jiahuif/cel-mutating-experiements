package api

type MutatingAdmissionPolicy struct {
	Mutation []struct {
		Expressions []string
	}
}

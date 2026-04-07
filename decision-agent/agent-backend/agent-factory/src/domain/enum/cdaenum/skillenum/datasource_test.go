package skillenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatasource_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Datasource("inherit_main"), DatasourceInheritMain)
	assert.Equal(t, Datasource("self_configured"), DatasourceSelfConfigured)
}

func TestDatasource_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []Datasource{
		DatasourceInheritMain,
		DatasourceSelfConfigured,
	}

	for _, ds := range validTypes {
		t.Run(string(ds), func(t *testing.T) {
			t.Parallel()

			err := ds.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestDatasource_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := Datasource("invalid_datasource")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid skill agent datasource")
}

func TestDatasource_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := Datasource("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
}

func TestDatasourceSpecificInherit_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DatasourceSpecificInherit("docs_only"), DatasourceInheritDocs)
	assert.Equal(t, DatasourceSpecificInherit("graph_only"), DatasourceInheritGraph)
	assert.Equal(t, DatasourceSpecificInherit("all"), DatasourceInheritAll)
}

func TestDatasourceSpecificInherit_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []DatasourceSpecificInherit{
		DatasourceInheritDocs,
		DatasourceInheritGraph,
		DatasourceInheritAll,
	}

	for _, ds := range validTypes {
		t.Run(string(ds), func(t *testing.T) {
			t.Parallel()

			err := ds.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestDatasourceSpecificInherit_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := DatasourceSpecificInherit("invalid_inherit")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据源继承类型不合法")
}

func TestDatasourceSpecificInherit_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := DatasourceSpecificInherit("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
}

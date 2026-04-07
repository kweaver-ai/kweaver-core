package dapo

import "testing"

func TestBizDomainAgentRelPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &BizDomainAgentRelPo{}
		tableName := po.TableName()

		expected := "t_biz_domain_agent_rel"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestBizDomainAgentTplRelPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &BizDomainAgentTplRelPo{}
		tableName := po.TableName()

		expected := "t_biz_domain_agent_tpl_rel"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

package upgrades

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/tendermint"
	"main/pkg/metrics"
	"main/pkg/query_info"
	"main/pkg/utils"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Querier struct {
	QueryUpgrades bool
	Logger        zerolog.Logger
	Cosmovisor    *cosmovisorPkg.Cosmovisor
	Tendermint    *tendermint.RPC
	Tracer        trace.Tracer
}

func NewQuerier(
	queryUpgrades bool,
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tendermint *tendermint.RPC,
	tracer trace.Tracer,
) *Querier {
	return &Querier{
		QueryUpgrades: queryUpgrades,
		Logger:        logger.With().Str("component", "upgrades_querier").Logger(),
		Cosmovisor:    cosmovisor,
		Tendermint:    tendermint,
		Tracer:        tracer,
	}
}

func (u *Querier) Enabled() bool {
	return u.Tendermint != nil && u.QueryUpgrades
}

func (u *Querier) Name() string {
	return "upgrades-querier"
}

func (u *Querier) Get(ctx context.Context) ([]metrics.MetricInfo, []query_info.QueryInfo) {
	childCtx, span := u.Tracer.Start(
		ctx,
		"Querier "+u.Name(),
		trace.WithAttributes(attribute.String("node", u.Name())),
	)
	defer span.End()

	upgrade, upgradePlanQuery, err := u.Tendermint.GetUpgradePlan(childCtx)
	if err != nil {
		u.Logger.Err(err).Msg("Could not get latest upgrade plan from Tendermint")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{upgradePlanQuery}
	}

	isUpgradePresent := upgrade != nil

	metricInfos := []metrics.MetricInfo{{
		MetricName: metrics.MetricNameUpgradeComing,
		Labels:     map[string]string{},
		Value:      utils.BoolToFloat64(isUpgradePresent),
	}}
	queryInfos := []query_info.QueryInfo{upgradePlanQuery}

	if !isUpgradePresent {
		return metricInfos, queryInfos
	}

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeInfo,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      utils.BoolToFloat64(isUpgradePresent),
	}, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeHeight,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      float64(upgrade.Height),
	})

	// Calculate upgrade estimated time
	upgradeTime, upgradeTimeQuery, err := u.Tendermint.GetEstimateBlockTime(childCtx, upgrade.Height)
	queryInfos = append(queryInfos, upgradeTimeQuery)

	if err != nil {
		u.Logger.Err(err).Msg("Could not get estimated upgrade time")
		return metricInfos, queryInfos
	}

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeEstimatedTime,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      float64(upgradeTime.Unix()),
	})

	if u.Cosmovisor == nil {
		u.Logger.Warn().
			Msg("Cosmovisor not initialized, not returning binary presence.")
		return metricInfos, queryInfos
	}

	upgrades, cosmovisorGetUpgradesQueryInfo, err := u.Cosmovisor.GetUpgrades(childCtx)
	if err != nil {
		u.Logger.Error().Err(err).Msg("Could not get Cosmovisor upgrades")
		queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)
		return metricInfos, queryInfos
	}

	queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)

	// From cosmovisor docs:
	// The name variable in upgrades/<name> is the lowercase URI-encoded name
	// of the upgrade as specified in the upgrade module plan.
	upgradeName := strings.ToLower(upgrade.Name)
	upgradeName = url.QueryEscape(upgradeName)

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeBinaryPresent,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      utils.BoolToFloat64(upgrades.HasUpgrade(upgradeName)),
	})

	return metricInfos, queryInfos
}

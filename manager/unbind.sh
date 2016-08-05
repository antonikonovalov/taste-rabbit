echo http://localhost:4567/unbind?from="$1""&to=""$2"
curl -XGET http://localhost:4567/unbind?from="$1""&to=""$2"
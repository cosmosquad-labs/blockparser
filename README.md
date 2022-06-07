# BlockParser

## Install, build

```bash
# install
go install

# or build
go build
```

## How to
```bash
# Usage : blockparser [chain-dir] [start-height] [end-height] [search-string]
blockparser ~/.crescent 402001 600000 cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e
```

output 
```
Loaded :  /Users/guest/.crescent/data/
Input Start Height : 402001
Input End Height : 432001
Latest Height : 620000

491419 [beginblock] type:"transfer" attributes:<key:"recipient" value:"cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e" index:true > attributes:<key:"sender" value:"cre10d07y265gmmuvt4z0w9aw880jnsr700j72qqr7" index:true > attributes:<key:"amount" value:"500000000ucre" index:true >
491419 [endblock] type:"coin_received" attributes:<key:"receiver" value:"cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e" index:true > attributes:<key:"amount" value:"500000000ucre" index:true >
491419 [endblock] type:"transfer" attributes:<key:"recipient" value:"cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e" index:true > attributes:<key:"sender" value:"cre10d07y265gmmuvt4z0w9aw880jnsr700j72qqr7" index:true > attributes:<key:"amount" value:"500000000ucre" index:true >
502310 [txs] [{"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"cre1hvpzhd8jx7lgyfla9t4yz4exqmxmmcraq65tue"},{"key":"amount","value":"33859739390000ucre"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e"},{"key":"amount","value":"33859739390000ucre"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.vesting.v1beta1.MsgCreatePeriodicVestingAccount"},{"key":"sender","value":"cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e"},{"key":"module","value":"vesting"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"cre1hvpzhd8jx7lgyfla9t4yz4exqmxmmcraq65tue"},{"key":"sender","value":"cre1u9jxn6l7seq5jjej4w6etpdxufphwfuunljr4e"},{"key":"amount","value":"33859739390000ucre"}]}]}]
...
```

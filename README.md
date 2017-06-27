# cryptom

Cryptocurrency Orders Monitor

## Supported exchanges / systems

Currently only Kraken exchange on the OSX operating system is supported.

## Setup

You first need to create an API key that allows querying open/closed
orders. In order to do this go to  your
[Kraken API page](https://www.kraken.com/u/settings/api) and
`Generate New Key` with the `Query Open Orders & Trades` permission.

Add the generated Key and Secret to a file called `.cryptom.toml` in
your home directory (replace *KEY / SECRET* with the values from
previous step):

```bash
echo "" >> ~/.cryptom.toml
echo kraken-key=\"KEY\" >> ~/.cryptom.toml
echo kraken-secret=\"SECRET\" >> ~/.cryptom.toml
```

Now run cryptom:

```bash
# run in foreground
./cryptom

# or you can send it to background
./cryptom &
```
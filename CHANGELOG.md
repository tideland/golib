# Tideland Go Library

## 2016-10-04

- Added *SplitFilter()* and *SplitMap()* to *stringex* for convenient splitting
  operations
- *Etc.Apply()* is now more robust by using *SplitMap()*

## 2016-10-03

- Added *Root()* to the trees in *collections*

## 2016-10-02

- *KeyValueTree* and *KeyStringValueTree* in *collections* now can copy subtrees
- Both now also support the retrieval and setting of keys
- Added *Split()* to *Etc*

## 2016-09-29

- Added *etc* based on former *configuration* package
- Marked *configuration* as deprecated
- Marked *web* as deprecated after migration to https://github.com/tideland/gorest

## 2016-02-16

- Removal of the *cells* package after migration to https://github.com/tideland/gocells

## 2015-09-01

- Added filtering to *logger* package

## 2015-08-23

- The backend of the *monitoring* package is now pluggable
- Beside the standard backend a null backend doing nothing has been added
- So last changes in *cells* packge have been rolled back as the monitoring
  handling is now a global topic

## 2015-08-18

- Monitoring in *cells* package is now pluggable

## 2015-08-17

- Fixed race condition in *cells* package
- Optimised time handling in *cells* package

## 2015-08-09

- Added `Collect()` and `DoAll()` to *errors* package

## 2015-08-02

- Added `BeginOf()` and `EndOf()` to *timex* package

## 2015-08-01

- Added `Set` and `StringSet` to *collections* package
- Added `Retry()` to *timex* package

## 2015-07-28

- Added assertion `Retry()` to *audit* package

## 2015-07-26

- Added `CallbackBehavior` to *cells* package

## 2015-07-24

- Fixed *cells* package unsubscribing failure when stopping cell with
  bi-directional subscriptions; thanks to Jonathan Camp for
  his fix
- Added expected value to compare with signal in `Wait()` assertion
- Added test for configuration validation in configurator behavior

## 2015-07-23

- Added `ReadFile()` to *configuration* package
- Added `SimpleProcessorBehavior` to *cells* package
- Added `ConfiguratorBehavior` to *cells* package
- Added assertion `Wait()` to *audit* package

## 2015-07-20

- Simplified *configuration* package for usage with `stringex.Defaulter`

## 2015-07-17

- Added *stringex* package

## 2015-07-10

- Added `KeyStringValueTreeBuilder` to *sml* package
- Several minor fixes

## 2015-07-05

- Made time format in *logger* package standard backend changeable

## 2015-06-28

- Changed *configuration* package to be more powerful
  and convenient

## 2015-06-25

- Added new `SceneBehavior` to *cells* package

## 2015-06-25

- Done migration into new library
- Added new *configuration* package

## 2015-06-05

- Started migration of existing packages into new library

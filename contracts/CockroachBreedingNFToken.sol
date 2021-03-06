pragma solidity ^0.4.20;

import "./CockroachNFToken.sol";


/**
 * @title CockroachBreedingNFToken
 *
 * Superset of the ERC721 standard that allows for the minting
 * of non-fungible tokens.
 */
contract CockroachBreedingNFToken is CockroachNFToken {
    using SafeMath for uint;

    uint256 public speedUnitFee = 1 finney;

    event Spawn(
        address indexed _to,
        uint256 _tokenId
    );

    function setSpeedUnitFee(uint256 val) external onlyCFO {
        speedUnitFee = val;
    }

    function calcSpeedPrice(uint8 _speed) public view returns (uint256) {
        return _speed * speedUnitFee;
    }

    function spawn(string _name, uint8 _speed) public payable {
        uint256 _unique = block.timestamp - block.coinbase.balance;
        spawnUnique(_name, _speed, _unique);
    }

    function spawnUnique(string _name, uint8 _speed, uint256 _unique) public payable {
        uint reqSpeedPrice = calcSpeedPrice(_speed);
        require(msg.value >= reqSpeedPrice);

        Cockroach memory cr = Cockroach({name : _name, speed : _speed, unique : _unique});
        uint256 newCockroachId = cockroaches.push(cr) - 1;
        _spawn(msg.sender, newCockroachId);

        // refund
        if (msg.value > reqSpeedPrice) {
            msg.sender.transfer(msg.value - reqSpeedPrice);
        }
    }

    function _spawn(address _owner, uint256 _tokenId) internal onlyNonexistentToken(_tokenId) {
        _setTokenOwner(_tokenId, _owner);
        _addTokenToOwnersList(_owner, _tokenId);
        population = population.add(1);
        Spawn(_owner, _tokenId);
    }
}

/**
 * @title SafeMath
 * @dev Math operations with safety checks that throw on error
 */
library SafeMath {

    /**
    * @dev Multiplies two numbers, throws on overflow.
    */
    function mul(uint256 a, uint256 b) internal pure returns (uint256) {
        if (a == 0) {
            return 0;
        }
        uint256 c = a * b;
        assert(c / a == b);
        return c;
    }

    /**
    * @dev Integer division of two numbers, truncating the quotient.
    */
    function div(uint256 a, uint256 b) internal pure returns (uint256) {
        // assert(b > 0); // Solidity automatically throws when dividing by 0
        uint256 c = a / b;
        // assert(a == b * c + a % b); // There is no case in which this doesn't hold
        return c;
    }

    /**
    * @dev Substracts two numbers, throws on overflow (i.e. if subtrahend is greater than minuend).
    */
    function sub(uint256 a, uint256 b) internal pure returns (uint256) {
        assert(b <= a);
        return a - b;
    }

    /**
    * @dev Adds two numbers, throws on overflow.
    */
    function add(uint256 a, uint256 b) internal pure returns (uint256) {
        uint256 c = a + b;
        assert(c >= a);
        return c;
    }
}
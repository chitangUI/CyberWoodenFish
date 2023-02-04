using System;
using System.Collections;
using System.Collections.Generic;
using Unity.VisualScripting;
using UnityEngine;
using UnityEngine.Serialization;
using UnityEngine.UI;

namespace ElectronicFish
{
	public class MainManager : MonoBehaviour
	{
		[SerializeField] private Button testAddButton;
		[SerializeField] private Text meritText;
		private List<Transform> _transforms;
		private GameObject _meritText;
		private Vector3 _targetVector;

		// Start is called before the first frame update
		private void Awake()
		{
			_meritText = meritText.GameObject();

			var position = _meritText.transform.position;
			position.y += 300;
			_targetVector = position;

			testAddButton.onClick.AddListener(MeritAdd);
		}

		private void Update()
		{
			foreach (var i in _transforms)
			{
				i.position = Vector3.Lerp(i.position, _targetVector, 0.4f);
			}
		}

		private void MeritAdd()
		{
			var newText = Instantiate(_meritText).GetComponent<Text>();
			newText.enabled = true;
			_transforms.Add(newText.transform);
		}
	}
}
